// The MIT License (MIT)
//
// Copyright © 2023 Tosone <i@tosone.cn>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

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
