// Copyright 2024 sigma
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

package etag

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"hash/crc32"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/consts"
)

// Skipper defines a function to skip middleware.
type Skipper func(c *gin.Context) bool

// defaultSkipper always returns false (never skips).
func defaultSkipper(c *gin.Context) bool { return false }

// EtagConfig defines the config for Etag middleware.
type EtagConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper
	// Weak defines if the Etag is weak or strong.
	Weak bool
	// HashFn defines the hash function to use. Default is crc32q.
	HashFn func(config EtagConfig) hash.Hash
}

var (
	// DefaultEtagConfig is the default Etag middleware config.
	DefaultEtagConfig = EtagConfig{
		Skipper: defaultSkipper,
		Weak:    true,
		HashFn: func(config EtagConfig) hash.Hash {
			if config.Weak {
				const crcPol = 0xD5828281
				crc32qTable := crc32.MakeTable(crcPol)
				return crc32.New(crc32qTable)
			}
			return sha256.New()
		},
	}
	normalizedETagName        = consts.HeaderETag
	normalizedIfNoneMatchName = consts.HeaderIfNoneMatch
	weakPrefix                = "W/"
)

// Etag returns a Etag middleware.
func Etag() gin.HandlerFunc {
	c := DefaultEtagConfig
	return WithEtagConfig(c)
}

// WithEtagConfig returns a Etag middleware with config.
func WithEtagConfig(config EtagConfig) gin.HandlerFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultEtagConfig.Skipper
	}

	return func(c *gin.Context) {
		skipper := config.Skipper
		if skipper == nil {
			skipper = DefaultEtagConfig.Skipper
		}

		if skipper(c) {
			c.Next()
			return
		}

		// get the hash function
		hashFn := config.HashFn
		if hashFn == nil {
			hashFn = DefaultEtagConfig.HashFn
		}

		originalWriter := c.Writer
		req := c.Request
		hw := bufferedWriter{ResponseWriter: originalWriter, hash: hashFn(config), buf: bytes.NewBuffer(nil)}
		c.Writer = &hw
		c.Next()
		// restore the original writer
		c.Writer = originalWriter

		resHeader := originalWriter.Header()

		if hw.hash == nil ||
			resHeader.Get(normalizedETagName) != "" ||
			strconv.Itoa(hw.status)[0] != '2' ||
			hw.status == http.StatusNoContent ||
			hw.buf.Len() == 0 {
			writeRaw(originalWriter, hw)
			return
		}

		etag := fmt.Sprintf("%v-%v", strconv.Itoa(hw.len),
			hex.EncodeToString(hw.hash.Sum(nil)))

		if config.Weak {
			etag = weakPrefix + etag
		}

		resHeader.Set(normalizedETagName, etag)

		ifNoneMatch := req.Header.Get(normalizedIfNoneMatchName) // get the If-None-Match header
		headerFresh := ifNoneMatch == etag || ifNoneMatch == weakPrefix+etag

		if headerFresh {
			originalWriter.WriteHeader(http.StatusNotModified)
			originalWriter.Write(nil) // nolint: errcheck
		} else {
			writeRaw(originalWriter, hw)
		}
	}
}

// bufferedWriter is a wrapper around gin.ResponseWriter that buffers the
// response and calculates the hash of the response. It embeds
// gin.ResponseWriter so the full interface is satisfied; only Write,
// WriteHeader and friends are overridden to intercept writes.
type bufferedWriter struct {
	gin.ResponseWriter
	hash    hash.Hash
	buf     *bytes.Buffer
	len     int
	status  int
	written bool
}

// WriteHeader sends an HTTP response header with the provided status code.
func (hw *bufferedWriter) WriteHeader(status int) {
	hw.status = status
}

// WriteHeaderNow is a no-op: headers are written after buffering completes.
func (hw *bufferedWriter) WriteHeaderNow() {}

// Write writes the data to the buffer to be sent as part of an HTTP reply.
func (hw *bufferedWriter) Write(b []byte) (int, error) {
	if hw.status == 0 {
		hw.status = http.StatusOK
	}
	hw.written = true
	// write to the buffer
	l, err := hw.buf.Write(b)
	if err != nil {
		return l, err
	}
	// write to the hash
	l, err = hw.hash.Write(b)
	hw.len += l
	return l, err
}

// WriteString writes a string to the buffer (goes through Write).
func (hw *bufferedWriter) WriteString(s string) (int, error) {
	return hw.Write([]byte(s))
}

// Status returns the buffered HTTP status code.
func (hw *bufferedWriter) Status() int {
	return hw.status
}

// Size returns the number of bytes written to the buffer.
func (hw *bufferedWriter) Size() int {
	return hw.len
}

// Written returns whether Write was called.
func (hw *bufferedWriter) Written() bool {
	return hw.written
}

// Flush is a no-op: the buffer is flushed after the handler returns.
func (hw *bufferedWriter) Flush() {}

// WriteTo writes the buffered data to the underlying gin.ResponseWriter.
func writeRaw(res gin.ResponseWriter, hw bufferedWriter) {
	res.WriteHeader(hw.status)
	res.Write(hw.buf.Bytes()) // nolint: errcheck
}
