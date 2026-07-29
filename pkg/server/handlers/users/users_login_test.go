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

package users

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	svcuser "github.com/go-sigma/sigma/pkg/service/users"
)

func TestLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := svcuser.NewMockUserService(ctrl)
	service.EXPECT().Login(gomock.Any(), "user-1", time.Minute, time.Hour).
		Return("access-token", "refresh-token", nil)
	recorder, c := newUserContext()
	email := "sigma@example.test"
	c.Set(consts.ContextUser, &models.User{ID: "user-1", Username: "sigma", Email: &email})

	(&handler{
		Config: &config.Configuration{Auth: config.ConfigurationAuth{Jwt: config.ConfigurationAuthJwt{
			Ttl:        time.Minute,
			RefreshTTL: time.Hour,
		}}},
		UserSvc: service,
	}).Login(c)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"token":"access-token"`)
	require.Contains(t, recorder.Body.String(), `"refresh_token":"refresh-token"`)
	require.Contains(t, recorder.Body.String(), `"username":"sigma"`)
}

func TestLoginWithoutUser(t *testing.T) {
	recorder, c := newUserContext()

	(&handler{
		Config:  &config.Configuration{},
		UserSvc: nil,
	}).Login(c)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func newUserContext() (*httptest.ResponseRecorder, *gin.Context) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	return recorder, c
}
