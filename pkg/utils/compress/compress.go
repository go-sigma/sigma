// Copyright 2023 XImager
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

package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"

	"github.com/rs/zerolog/log"
)

// Compress compresses the given string using gzip.
func Compress(src string) ([]byte, error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := srcFile.Close()
		if err != nil {
			log.Warn().Err(err).Msg("Close file failed")
		}
	}()

	var dst bytes.Buffer
	gzipWriter, err := gzip.NewWriterLevel(&dst, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(gzipWriter, srcFile)
	if err != nil {
		return nil, err
	}

	err = gzipWriter.Close()
	if err != nil {
		return nil, err
	}

	return dst.Bytes(), nil
}
