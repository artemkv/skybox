package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// reconciliation

type LocalObjectInfo struct {
	Path  string
	Size  int64
	Hash  string
	Error error
}

type CloudObjectInfo struct {
	Path string
	Size int64
	Hash string
}

func computeFileHash(folder string, file *LocalFileInfo, hashCache *HashCache, hashCacheKey string) (string, error) {
	hash, ok := hashCache.HashLookup[hashCacheKey]
	if ok {
		return hash, nil
	}
	return GetHash(filepath.Join(folder, file.Path))
}

func toLocalObjectList(folder string, files []*LocalFileInfo, hashCache *HashCache) ([]*LocalObjectInfo, *HashCache) {
	var objects []*LocalObjectInfo

	// we just re-build cache from scratch every time
	newCache := &HashCache{
		HashLookup: make(map[string]string, len(files)),
	}

	for _, file := range files {
		// if already has error, simply wrap
		if file.Error != nil {
			objInfo := &LocalObjectInfo{
				Path:  file.Path,
				Error: file.Error,
			}
			objects = append(objects, objInfo)
			continue
		}

		// get cache
		hashCacheKey := GetHashCacheKey(folder, file)
		hash, err := computeFileHash(folder, file, hashCache, hashCacheKey)

		// error getting hash
		if err != nil {
			objInfo := &LocalObjectInfo{
				Path:  file.Path,
				Error: fmt.Errorf("err reading file: %w", err),
			}
			objects = append(objects, objInfo)
			continue
		}

		// got hash
		objInfo := &LocalObjectInfo{
			Path: file.Path,
			Size: file.Size,
			Hash: hash,
		}
		objects = append(objects, objInfo)
		newCache.HashLookup[hashCacheKey] = hash
	}

	return objects, newCache
}

func toCloudObjectList(folderItems []*FolderMetaItem) []*CloudObjectInfo {
	var objects []*CloudObjectInfo

	for _, item := range folderItems {
		objInfo := &CloudObjectInfo{
			Path: item.Path,
			Size: item.Size,
			Hash: item.Hash,
		}
		objects = append(objects, objInfo)
	}

	return objects
}

// missing == in backup
// Skips objects with errors
// Skips empty objects
func getMissingObjects(local []*LocalObjectInfo, cloud []*CloudObjectInfo) []*LocalObjectInfo {
	hashLookup := make(map[string]int)
	for _, obj := range cloud {
		hashLookup[obj.Hash] = 1
	}

	var objects []*LocalObjectInfo
	for _, obj := range local {
		if obj.Error != nil {
			continue
		}
		if obj.Size == 0 {
			continue
		}

		if _, ok := hashLookup[obj.Hash]; !ok {
			objects = append(objects, obj)
		}
	}
	return objects
}

func hasErrors(local []*LocalObjectInfo) bool {
	for _, obj := range local {
		if obj.Error != nil {
			return true
		}
	}
	return false
}

// orphaned == in backup
func getOrphanedObjects(local []*LocalObjectInfo, cloud []*CloudObjectInfo) []*CloudObjectInfo {
	hashLookup := make(map[string]int)
	for _, obj := range local {
		hashLookup[obj.Hash] = 1
	}

	var objects []*CloudObjectInfo
	for _, obj := range cloud {
		if _, ok := hashLookup[obj.Hash]; !ok {
			objects = append(objects, obj)
		}
	}
	return objects
}

// make keys

func makeFolderMetaKey(deviceId string) string {
	return fmt.Sprintf("%s/_folder.meta", deviceId)
}

func makeObjectKey(deviceId string, hash string) string {
	return fmt.Sprintf("%s/%s", deviceId, hash)
}

// metadata serialization

const (
	FileNonceMetaKey                  = "nonce"
	FileEncryptionKeyEncryptedMetaKey = "filekey"
	FileEncryptionKeyNonceMetaKey     = "filekeynonce"
)

type FolderMeta struct {
	Items []*FolderMetaItem `json:"items"`
}

type FolderMetaItem struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
	Hash string `json:"hash"`
}

func toJson(obj any) (string, error) {
	// serialize message
	bytes, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func toFolderMeta(data []byte) (*FolderMeta, error) {
	var meta FolderMeta
	err := json.Unmarshal(data, &meta)
	if err != nil {
		return nil, err
	}
	return &meta, nil
}

func toFolderMetaItems(local []*LocalObjectInfo, prevFolderItems []*FolderMetaItem) []*FolderMetaItem {
	prevLookup := make(map[string]*FolderMetaItem)
	for _, item := range prevFolderItems {
		prevLookup[item.Path] = item
	}

	var items []*FolderMetaItem = make([]*FolderMetaItem, 0, len(local))
	for _, obj := range local {
		if obj.Error != nil {
			prev, ok := prevLookup[obj.Path]
			if ok {
				items = append(items, prev)
			} else {
				// Ignore
			}
		} else {
			itemInfo := &FolderMetaItem{
				Path: obj.Path,
				Size: obj.Size,
				Hash: obj.Hash,
			}
			items = append(items, itemInfo)
		}
	}
	return items
}

// upload/download

func uploadObjectEncrypted(
	bucket string,
	folder string,
	obj *LocalObjectInfo,
	objectKey string,
	masterKey []byte) error {
	// get new unique ecryption key and nonce for a file
	fileEncryptionKey := GenerateNewEncryptionKey()
	fileNonce := GenerateNewNonce()

	// Encode the file encryption key with master key
	fileEncryptionKeyNonce := GenerateNewNonce()
	var fileEncryptionKeyEncrypted bytes.Buffer
	err := Encrypt(
		bytes.NewReader(fileEncryptionKey),
		&fileEncryptionKeyEncrypted,
		masterKey,
		fileEncryptionKeyNonce)
	if err != nil {
		return err
	}

	// prepare file meta
	meta := map[string]string{
		FileNonceMetaKey:                  base64.StdEncoding.EncodeToString(fileNonce),
		FileEncryptionKeyEncryptedMetaKey: base64.StdEncoding.EncodeToString(fileEncryptionKeyEncrypted.Bytes()),
		FileEncryptionKeyNonceMetaKey:     base64.StdEncoding.EncodeToString(fileEncryptionKeyNonce),
	}

	// open the file
	input, err := os.Open(filepath.Join(folder, obj.Path))
	if err != nil {
		return err
	}
	defer input.Close()

	// set up piped encryption
	pipeReader, pipeWriter := io.Pipe()
	go func() {
		defer pipeWriter.Close()

		err := Encrypt(input, pipeWriter, fileEncryptionKey, fileNonce)
		if err != nil {
			pipeWriter.CloseWithError(err)
			return
		}
	}()

	// stream to s3 with metadata
	err = UploadFile(bucket, objectKey, pipeReader, meta)
	if err != nil {
		return err
	}

	return nil
}

func uploadFolderMeta(bucket string, folderMetaKey string, meta *FolderMeta) error {
	metaJson, err := toJson(meta)
	if err != nil {
		return err
	}

	err = UploadFile(bucket, folderMetaKey, strings.NewReader(metaJson), map[string]string{})
	if err != nil {
		return err
	}

	return nil
}

// Resolves not found, and provides empty metadata in that case
func downloadFolderMeta(bucket string, folderMetaKey string) (*FolderMeta, error) {
	json, err := GetFileContent(bucket, folderMetaKey)
	if err != nil {
		return nil, err
	}

	if json == nil {
		var items []*FolderMetaItem = make([]*FolderMetaItem, 0)
		return &FolderMeta{
			Items: items,
		}, err
	}

	folderMeta, err := toFolderMeta(json)
	if err != nil {
		return nil, err
	}

	return folderMeta, nil
}

func Backup(folder string, bucket string, deviceId string, masterKey []byte) ([]*LocalObjectInfo, error) {
	// get cached hashes
	hashCache := RestoreHashes()

	// get local files
	fmt.Println("Loading local folder")
	files, err := ListFolder(folder)
	if err != nil {
		return nil, err
	}
	localObjects, updHashes := toLocalObjectList(folder, files, hashCache)
	err = StoreHashes(updHashes)
	if err != nil {
		fmt.Printf("Failed to store hash cache: %v\n", err)
	}
	fmt.Println("Loading local folder done")

	// get folder metadata
	fmt.Println("Retrieving cloud folder metadata")
	folderMetaKey := makeFolderMetaKey(deviceId)
	folderMeta, err := downloadFolderMeta(bucket, folderMetaKey)
	if err != nil {
		return localObjects, err
	}
	cloudObjects := toCloudObjectList(folderMeta.Items)
	fmt.Println("Retrieving cloud folder metadata done")

	// upload all missing blobs
	// (only for files that have no errors)
	fmt.Println("Uploading missing objects")
	missing := getMissingObjects(localObjects, cloudObjects)
	for _, obj := range missing {
		fmt.Printf("Uploading content for: '%s'\n", obj.Path)
		objectKey := makeObjectKey(deviceId, obj.Hash)
		err = uploadObjectEncrypted(bucket, folder, obj, objectKey, masterKey)
		if err != nil {
			obj.Error = fmt.Errorf("failed to upload: %v", err)
		}
	}
	fmt.Println("Uploading missing objects done")

	// remove orphans
	fmt.Println("Removing orphans")
	if !hasErrors(localObjects) {
		orphans := getOrphanedObjects(localObjects, cloudObjects)
		for _, obj := range orphans {
			fmt.Printf("Removing: '%s'\n", obj.Path)
			objectKey := makeObjectKey(deviceId, obj.Hash)
			err = DeleteFile(bucket, objectKey)
			if err != nil {
				fmt.Printf("Failed to delete the object: %v\n", err)
			}
		}
	}
	fmt.Println("Removing orphans done")

	// re-create and upload new metadata file
	// (only for files that have no errors)
	fmt.Println("Saving folder metadata")
	newFolderMeta := &FolderMeta{
		Items: toFolderMetaItems(localObjects, folderMeta.Items),
	}
	err = uploadFolderMeta(bucket, folderMetaKey, newFolderMeta)
	if err != nil {
		return localObjects, err
	}
	fmt.Println("Saving folder metadata done")

	return localObjects, nil
}
