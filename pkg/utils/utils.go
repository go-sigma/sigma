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

package utils

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// GetContentLength returns the content length of the request.
func GetContentLength(req *http.Request) (int64, error) {
	if req == nil {
		return 0, fmt.Errorf("request is nil")
	}
	str := req.Header.Get("Content-Length")
	if str == "" {
		return 0, nil
	}
	length, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("content length is not a number")
	}
	return length, nil
}

// GenPathByDigest generates the path by the digest.
func GenPathByDigest(digest digest.Digest) string {
	hex := digest.Hex()
	return fmt.Sprintf("%s/%s/%s/%s", digest.Algorithm(), hex[0:2], hex[2:4], hex[4:])
}

// SetLevel sets the log level
func SetLevel(level int) {
	if level < int(zerolog.TraceLevel) || level > int(zerolog.FatalLevel) {
		level = int(zerolog.InfoLevel)
	}

	var timeFormat = "2006-01-02 15:04:05" // change it to 'time.DataTime' om go 1.20
	zerolog.SetGlobalLevel(zerolog.Level(level))
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: timeFormat, FormatCaller: func(i interface{}) string {
		var c string
		if cc, ok := i.(string); ok {
			c = cc
		}
		if len(c) > 0 && strings.Contains(c, "/") {
			lastIndex := strings.LastIndex(c, "/")
			c = c[lastIndex+1:]
		}
		return c
	}}).With().Caller().Timestamp().Logger()
}
