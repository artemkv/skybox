package main

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/chacha20"
	"golang.org/x/crypto/scrypt"
)

const KeyLength = 32

func DeriveMasterKey(secret string, salt []byte) ([]byte, error) {
	return scrypt.Key([]byte(secret), salt, 32768, 8, 1, KeyLength)
}

func GenerateNewEncryptionKey() []byte {
	key := make([]byte, KeyLength)
	rand.Read(key)
	return key
}

func GenerateNewNonce() []byte {
	nonce := make([]byte, chacha20.NonceSizeX)
	rand.Read(nonce)
	return nonce
}

func Encrypt(reader io.Reader, writer io.Writer, key []byte, nonce []byte) error {
	cipherStream, err := chacha20.NewUnauthenticatedCipher(key, nonce)
	if err != nil {
		return fmt.Errorf("failed to create xchacha20 cipher: %w", err)
	}

	wrapper := &cipher.StreamWriter{
		S: cipherStream,
		W: writer,
	}

	if _, err := io.Copy(wrapper, reader); err != nil {
		return fmt.Errorf("failed to pipe to encrypted stream: %w", err)
	}

	return nil
}

func Decrypt(reader io.Reader, writer io.Writer, key []byte, nonce []byte) error {
	cipherStream, err := chacha20.NewUnauthenticatedCipher(key, nonce)
	if err != nil {
		return fmt.Errorf("failed to create xchacha20 cipher: %w", err)
	}

	wrapper := &cipher.StreamReader{
		S: cipherStream,
		R: reader,
	}

	if _, err := io.Copy(writer, wrapper); err != nil {
		return fmt.Errorf("failed to pipe to decrypted stream: %w", err)
	}

	return nil
}
