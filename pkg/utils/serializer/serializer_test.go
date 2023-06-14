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

package serializer

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	testify "github.com/stretchr/testify/assert"
)

type (
	user struct {
		ID   int    `json:"id" xml:"id" form:"id" query:"id" param:"id" header:"id"`
		Name string `json:"name" xml:"name" form:"name" query:"name" param:"name" header:"name"`
	}
)

const (
	userJSON       = `{"id":1,"name":"Jon Snow"}`
	invalidContent = "invalid content"
)

const userJSONPretty = `{
  "id": 1,
  "name": "Jon Snow"
}`

// see here: https://github.com/labstack/echo/blob/master/json_test.go

// Note this test is deliberately simple as there's not a lot to test.
// Just need to ensure it writes JSONs. The heavy work is done by the context methods.
func TestDefaultJSONCodec_Encode(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert := testify.New(t)

	// Echo
	assert.Equal(e, c.Echo())

	// Request
	assert.NotNil(c.Request())

	// Response
	assert.NotNil(c.Response())

	//--------
	// Default JSON encoder
	//--------

	enc := new(DefaultJSONSerializer)

	err := enc.Serialize(c, user{1, "Jon Snow"}, "")
	if assert.NoError(err) {
		assert.Equal(userJSON+"\n", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = enc.Serialize(c, user{1, "Jon Snow"}, "  ")
	if assert.NoError(err) {
		assert.Equal(userJSONPretty+"\n", rec.Body.String())
	}
}

// Note this test is deliberately simple as there's not a lot to test.
// Just need to ensure it writes JSONs. The heavy work is done by the context methods.
func TestDefaultJSONCodec_Decode(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert := testify.New(t)

	// Echo
	assert.Equal(e, c.Echo())

	// Request
	assert.NotNil(c.Request())

	// Response
	assert.NotNil(c.Response())

	//--------
	// Default JSON encoder
	//--------

	enc := new(DefaultJSONSerializer)

	var u = user{}
	err := enc.Deserialize(c, &u)
	if assert.NoError(err) {
		assert.Equal(u, user{ID: 1, Name: "Jon Snow"})
	}

	var userUnmarshalSyntaxError = user{}
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(invalidContent))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = enc.Deserialize(c, &userUnmarshalSyntaxError)
	assert.IsType(&echo.HTTPError{}, err)

	var userUnmarshalTypeError = struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}{}

	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = enc.Deserialize(c, &userUnmarshalTypeError)
	assert.IsType(&echo.HTTPError{}, err)
}
