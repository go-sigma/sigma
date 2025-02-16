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

package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// GetFromCtx ...
func GetFromCtx[T any](ctx echo.Context, key string) (T, bool) {
	val := ctx.Get(key)
	v, ok := val.(T)
	return v, ok
}

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

// BindValidate binds and validates the request body
func BindValidate(c echo.Context, data any) error {
	err := c.Bind(data)
	if err != nil {
		return err
	}
	err = c.Validate(data)
	if err != nil {
		return err
	}
	return nil
}

// PanicIf panics if err is not nil
func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

// Inject injects source to target if source is not nil
func Inject(target any, source any) error {
	if source == nil {
		return nil
	}
	return copier.Copy(target, source)
}

// NormalizePagination normalizes the pagination
func NormalizePagination(in types.Pagination) types.Pagination {
	if in.Page == nil || ptr.To(in.Page) < 1 {
		in.Page = ptr.Of(int(1))
	}
	if in.Limit == nil || ptr.To(in.Limit) > 100 || ptr.To(in.Limit) <= 0 {
		in.Limit = ptr.Of(int(10))
	}
	return in
}

// TrimHTTP ...
func TrimHTTP(in string) string {
	if strings.HasPrefix(in, "http://") {
		return strings.TrimPrefix(in, "http://")
	} else if strings.HasPrefix(in, "https://") {
		return strings.TrimPrefix(in, "https://")
	}
	return strings.TrimSuffix(in, "/")
}

// IsDir returns true if given path is a directory,
// or returns false when it's a file or does not exist.
func IsDir(dir string) bool {
	f, e := os.Stat(dir)
	if e != nil {
		return false
	}
	return f.IsDir()
}

// IsFile returns true if given path is a file,
// or returns false when it's a directory or does not exist.
func IsFile(filePath string) bool {
	f, e := os.Stat(filePath)
	if e != nil {
		return false
	}
	return !f.IsDir()
}

// IsExist checks whether a file or directory exists.
// It returns false when the file or directory does not exist.
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// MustMarshal marshals the given object to bytes.
func MustMarshal(in any) []byte {
	result, err := json.Marshal(in)
	if err != nil {
		panic(err)
	}
	return result
}

// DirWithSlash returns the dir with slash
func DirWithSlash(id string) string {
	// judge if the string has two slashes
	if len(strings.Split(id, "/")) == 3 {
		return id
	}
	if len(id) > 2 {
		if !strings.Contains(id, "/") {
			return DirWithSlash(fmt.Sprintf("%s/%s", id[0:2], id[2:]))
		}
		// remove the str before the last slash
		str := id[strings.LastIndex(id, "/")+1:]
		if len(str) > 2 {
			str = fmt.Sprintf("%s/%s", str[0:2], str[2:])
		} else {
			return fmt.Sprintf("%s/%s", strings.TrimSuffix(id[0:strings.LastIndex(id, "/")+1], "/"), str)
		}
		return DirWithSlash(fmt.Sprintf("%s/%s", strings.TrimSuffix(id[0:strings.LastIndex(id, "/")+1], "/"), str))
	}
	return id
}

type stringsJoin interface {
	String() string
}

// StringsJoin ...
func StringsJoin[T stringsJoin](strs []T, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0].String()
	}
	b := make([]string, len(strs))
	for i, str := range strs {
		b[i] = str.String()
	}
	return strings.Join(b, sep)
}

// UnwrapJoinedErrors ...
func UnwrapJoinedErrors(err error) string {
	e, ok := err.(interface{ Unwrap() []error })
	if !ok {
		return err.Error()
	}
	es := e.Unwrap()
	var ss = make([]string, len(es))
	for index, e := range es {
		ss[index] = e.Error()
	}
	return strings.Join(ss, ": ")
}

// GetUserFromCtx ...
func GetUserFromCtx(c echo.Context) (*models.User, bool, error) {
	iuser := c.Get(consts.ContextUser)
	if iuser == nil {
		log.Error().Msg("Get user from header failed")
		return nil, true, xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
	}
	user, ok := iuser.(*models.User)
	if !ok {
		log.Error().Msg("Convert user from header failed")
		return nil, true, xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
	}
	return user, false, nil
}

// GetUserFromCtxForDs ...
func GetUserFromCtxForDs(c echo.Context) (*models.User, bool, error) {
	iuser := c.Get(consts.ContextUser)
	if iuser == nil {
		log.Error().Msg("Get user from header failed")
		return nil, true, xerrors.NewDSError(c, xerrors.DSErrCodeUnauthorized)
	}
	user, ok := iuser.(*models.User)
	if !ok {
		log.Error().Msg("Convert user from header failed")
		return nil, true, xerrors.NewDSError(c, xerrors.DSErrCodeUnauthorized)
	}
	return user, false, nil
}

// OnceWithErr ...
func OnceWithErr(once *sync.Once, fn func() error) error {
	var errChan = make(chan error, 1)
	defer close(errChan)
	once.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("%v", r)
			}
		}()
		err := fn()
		errChan <- err
	})
	return <-errChan
}

// GetObjFromDigCon ...
func GetObjFromDigCon[T any](digCon *dig.Container) (T, error) {
	var result T
	err := digCon.Invoke(func(v T) {
		result = v
	})
	if err != nil {
		return result, err
	}
	return result, nil
}

// MustGetObjFromDigCon ...
func MustGetObjFromDigCon[T any](digCon *dig.Container) T {
	result, err := GetObjFromDigCon[T](digCon)
	if err != nil {
		panic(err)
	}
	return result
}

// GenRsaPriKey ...
func GenRsaPriKey(length int) (string, error) {
	if length != 1024 && length != 2048 && length != 4096 {
		return "", fmt.Errorf("rsa length is not allow")
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, length)
	if err != nil {
		return "", fmt.Errorf("generate rsa private key failed")
	}

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	return base64.StdEncoding.EncodeToString(pem.EncodeToMemory(privateKeyPEM)), nil
}
