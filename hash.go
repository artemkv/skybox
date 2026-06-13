package main

import (
	"fmt"
	"io"
	"os"

	"github.com/zeebo/blake3"
)

func GetHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := blake3.New()

	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	hash := hasher.Sum(nil)
	return fmt.Sprintf("%x", hash), nil
}
