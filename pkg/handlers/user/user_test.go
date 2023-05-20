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

package user

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/ximager/ximager/pkg/dal"
	"github.com/ximager/ximager/pkg/inits"
	"github.com/ximager/ximager/pkg/logger"
	"github.com/ximager/ximager/pkg/tests"
	"github.com/ximager/ximager/pkg/validators"
)

func TestFactory(t *testing.T) {
	logger.SetLevel("debug")
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	validators.Initialize(e)
	err := tests.Initialize()
	assert.NoError(t, err)
	err = tests.DB.Init()
	assert.NoError(t, err)
	defer func() {
		conn, err := dal.DB.DB()
		assert.NoError(t, err)
		err = conn.Close()
		assert.NoError(t, err)
		err = tests.DB.DeInit()
		assert.NoError(t, err)
	}()

	viper.SetDefault("auth.internalUser.password", "internal-ximager")
	viper.SetDefault("auth.internalUser.username", "internal-ximager")
	viper.SetDefault("auth.admin.password", "ximager")
	viper.SetDefault("auth.admin.username", "ximager")
	viper.SetDefault("auth.jwt.privateKey", privateKeyString)

	miniRedis := miniredis.RunT(t)
	viper.SetDefault("redis.url", "redis://"+miniRedis.Addr())

	err = inits.Initialize()
	assert.NoError(t, err)

	var f = factory{}
	err = f.Initialize(e)
	assert.NoError(t, err)

	go func() {
		err = e.Start(":8080")
		assert.ErrorIs(t, err, http.ErrServerClosed)
	}()

	time.Sleep(1 * time.Second)

	url := "http://127.0.0.1:8080/user/token"

	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	req.SetBasicAuth("ximager", "ximager")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	err = resp.Body.Close()
	assert.NoError(t, err)

	err = e.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestFactoryFailed(t *testing.T) {
	viper.SetDefault("auth.jwt.privateKey", privateKeyString+"1")
	var f = factory{}
	err := f.Initialize(echo.New())
	assert.Error(t, err)
}
