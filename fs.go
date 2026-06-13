package main

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"time"
)

const DefaultFilePermissions = 0644

// Size and ModTime can be null, in which case Error should be present
type LocalFileInfo struct {
	Path    string
	Size    int64
	ModTime time.Time
	Error   error
}

func ListFolder(folderPath string) ([]*LocalFileInfo, error) {
	var files []*LocalFileInfo

	err := filepath.Walk(folderPath,
		func(path string, info fs.FileInfo, err error) error {
			relPath, errOnRelPath := filepath.Rel(folderPath, path)
			if errOnRelPath != nil {
				// We cannot reliably continue if we can't detect relative path
				// Skipping these files would look like they don't exist
				log.Fatal(err)
			}

			if err != nil {
				fileInfo := &LocalFileInfo{
					Path:  filepath.ToSlash(relPath),
					Error: fmt.Errorf("err reading file stats: %w", err),
				}
				files = append(files, fileInfo)
				return nil
			}

			if !info.IsDir() {
				fileInfo := &LocalFileInfo{
					Path:    filepath.ToSlash(relPath),
					Size:    info.Size(),
					ModTime: info.ModTime(),
				}
				files = append(files, fileInfo)
			}
			return nil
		})

	if err != nil {
		return nil, err
	}
	return files, nil
}
