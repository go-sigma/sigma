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

package distribution

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/dig"
)

type factoryOk struct{}

func (f *factoryOk) Initialize(_ *gin.Context, _ *dig.Container) error {
	return nil
}

func TestInitializeOK(t *testing.T) {
	routerFactories = make([]Item, 0)
	err := RegisterRouterFactory(&factoryOk{}, 1)
	assert.NoError(t, err)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Content-Type", "application/json")
	err = All(c, dig.New())
	assert.NoError(t, err)
}

type factoryErr struct{}

func (f *factoryErr) Initialize(_ *gin.Context, _ *dig.Container) error {
	return errors.New("error")
}

func TestInitializeErr(t *testing.T) {
	routerFactories = make([]Item, 0)
	err := RegisterRouterFactory(&factoryErr{}, 1)
	assert.NoError(t, err)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Content-Type", "application/json")
	err = All(c, dig.New())
	assert.Error(t, err)
}

func TestInitializeDup(t *testing.T) {
	routerFactories = make([]Item, 0)
	err := RegisterRouterFactory(&factoryErr{}, 1)
	assert.NoError(t, err)
	err = RegisterRouterFactory(&factoryErr{}, 1)
	assert.Error(t, err)
}

type factoryContinue1 struct{}

func (f *factoryContinue1) Initialize(_ *gin.Context, _ *dig.Container) error {
	return ErrNext
}

type factoryContinue2 struct{}

func (f *factoryContinue2) Initialize(_ *gin.Context, _ *dig.Container) error {
	return ErrNext
}

func TestInitializeContinue(t *testing.T) {
	routerFactories = make([]Item, 0)
	err := RegisterRouterFactory(&factoryContinue1{}, 1)
	assert.NoError(t, err)
	err = RegisterRouterFactory(&factoryContinue2{}, 2)
	assert.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v2/test-none-exist", nil)
	c.Request.Header.Set("Content-Type", "application/json")
	err = All(c, dig.New())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}
