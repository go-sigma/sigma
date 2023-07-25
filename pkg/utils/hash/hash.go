// Copyright 2023 sigma
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

package hash

import (
	"crypto/md5"  // nolint: gosec
	"crypto/sha1" // nolint: gosec
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"os"
	"strings"
)

// ErrNoSuchFile ...
var ErrNoSuchFile = errors.New("no such file")

func isFile(filePath string) bool {
	f, e := os.Stat(filePath)
	if e != nil {
		return false
	}
	return !f.IsDir()
}

func selectMethod(method []string) hash.Hash {
	if len(method) == 1 {
		switch strings.ToLower(method[0]) {
		case "md5":
			return md5.New() // nolint: gosec
		case "sha1":
			return sha1.New() // nolint: gosec
		case "sha256":
			return sha256.New()
		case "sha512":
			return sha512.New()
		default:
			return sha256.New()
		}
	}
	return sha256.New()
}

// FileVerify 哈希校验一个文件，以 hex 的方式校验，method 支持：md5，sha1，sha256(默认)，sha512
func FileVerify(file, hash string, method ...string) (verify bool, err error) {
	if !isFile(file) {
		err = ErrNoSuchFile
		return
	}

	var fileObj *os.File
	if fileObj, err = os.Open(file); err != nil {
		return
	}
	var data = make([]byte, cacheSize)
	var n int
	var h = selectMethod(method)
	for {
		n, err = fileObj.Read(data)
		if n != 0 {
			if _, err = h.Write(data[:n]); err != nil {
				return
			}
		}
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			return
		}
	}
	verify = hex.EncodeToString(h.Sum(nil)) == hash
	return
}

// String 哈希一个字符串，输出 hex 的编码，method 支持：md5，sha1，sha256(默认)，sha512
func String(str string, method ...string) (hash string, err error) {
	var h = selectMethod(method)
	if _, err = h.Write([]byte(str)); err != nil {
		return
	}
	hash = hex.EncodeToString(h.Sum(nil))
	return
}

// File 哈希一个文件，输出 hex 的编码，method 支持：md5，sha1，sha256(默认)，sha512
func File(file string, method ...string) (hash string, err error) {
	if !isFile(file) {
		err = ErrNoSuchFile
		return
	}
	var fileObj *os.File
	if fileObj, err = os.Open(file); err != nil {
		return
	}
	var data = make([]byte, cacheSize)
	var n int
	var h = selectMethod(method)
	for {
		n, err = fileObj.Read(data)
		if n != 0 {
			if _, err = h.Write(data[:n]); err != nil {
				return
			}
		}
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			return
		}
	}
	hash = hex.EncodeToString(h.Sum(nil))
	return
}

const (
	cacheSize = 10240
)

// Reader ...
func Reader(reader io.Reader, method ...string) (hash string, err error) {
	var h = selectMethod(method)
	var data = make([]byte, cacheSize)
	var n int
	for {
		n, err = reader.Read(data)
		if n != 0 {
			if _, err = h.Write(data[:n]); err != nil {
				return
			}
		}
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			return
		}
	}
	hash = hex.EncodeToString(h.Sum(nil))
	return
}
