package crypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

// MustEncrypt ...
func MustEncrypt(key, plaintext string) string {
	result, err := Encrypt(key, plaintext)
	if err != nil {
		panic(fmt.Sprintf("encrypt string failed: %v", err))
	}
	return result
}

// Encrypt ...
func Encrypt(key, plaintext string) (string, error) {
	keyBytes := sha256.Sum256([]byte(key))

	block, err := aes.NewCipher(keyBytes[:])
	if err != nil {
		return "", err
	}

	iv := make([]byte, aes.BlockSize)
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return "", err
	}

	reader := &cipher.StreamReader{S: cipher.NewCFBEncrypter(block, iv), R: strings.NewReader(plaintext)}
	ciphertext, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.WithPadding(base64.StdPadding).EncodeToString(append(iv, ciphertext...)), nil
}

// Decrypt ...
func Decrypt(key, ciphertext string) (string, error) {
	keyBytes := sha256.Sum256([]byte(key))

	srcBytes, err := base64.StdEncoding.WithPadding(base64.StdPadding).DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	if len(srcBytes) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext should be have iv and length bigger than %d bytes", aes.BlockSize)
	}

	block, err := aes.NewCipher(keyBytes[:])
	if err != nil {
		return "", err
	}

	iv := srcBytes[:aes.BlockSize]

	reader := &cipher.StreamReader{S: cipher.NewCFBDecrypter(block, iv), R: bytes.NewReader(srcBytes[aes.BlockSize:])}

	plaintext, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
