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

package xerrors

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewDSError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.SetBasicAuth("sigma", "sigma1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := NewDSError(c, DSErrCodeBlobUnknown)
	assert.NoError(t, err)
}

func TestDSErrCode(t *testing.T) {
	e := ErrCode{Code: "title", Title: "message"}
	assert.Equal(t, "title: message", e.Error())
}
