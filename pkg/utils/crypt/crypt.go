// Copyright 2025 sigma
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

// MustSeal encrypts plaintext with key using AES-256-GCM and panics if encryption fails.
func MustSeal(key, plaintext string) string {
	result, err := Seal(key, plaintext)
	if err != nil {
		panic(fmt.Sprintf("seal string failed: %v", err))
	}
	return result
}

// Seal encrypts plaintext with AES-256-GCM and returns base64(nonce|ciphertext|tag).
//
// The result is authenticated: Open fails when the ciphertext is tampered with or
// the key differs. Use Seal for every value that leaves the process boundary, such
// as secrets stored at rest or passed to build runners via environment variables.
func Seal(key, plaintext string) (string, error) {
	keyBytes := sha256.Sum256([]byte(key))

	block, err := aes.NewCipher(keyBytes[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Open decrypts a base64(nonce|ciphertext|tag) value produced by Seal.
func Open(key, ciphertext string) (string, error) {
	keyBytes := sha256.Sum256([]byte(key))

	srcBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(keyBytes[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(srcBytes) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext should be have nonce of %d bytes", gcm.NonceSize())
	}

	nonce, data := srcBytes[:gcm.NonceSize()], srcBytes[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
