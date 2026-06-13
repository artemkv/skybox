package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const hashCacheFileName = ".hashes"

type HashCache struct {
	HashLookup map[string]string `json:"hashes"`
}

func StoreHashes(hashes *HashCache) error {
	hashesJson, err := toJson(hashes)
	if err != nil {
		return err
	}

	err = os.WriteFile(hashCacheFileName, []byte(hashesJson), DefaultFilePermissions)
	if err != nil {
		return err
	}

	return nil
}

func RestoreHashes() *HashCache {
	hashesJson, _ := os.ReadFile(hashCacheFileName)

	var hashes HashCache
	err := json.Unmarshal(hashesJson, &hashes)
	if err != nil {
		return &HashCache{}
	}

	return &hashes
}

func GetHashCacheKey(folder string, fileInfo *LocalFileInfo) string {
	path := filepath.Join(folder, fileInfo.Path)
	return fmt.Sprintf("%s|%d|%d", path, fileInfo.Size, fileInfo.ModTime.UnixNano())
}
