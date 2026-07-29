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
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal/models"
	svcuser "github.com/go-sigma/sigma/pkg/service/users"
)

func TestSignup(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := svcuser.NewMockUserService(ctrl)
	request := api.PostUserSignupRequest{
		Username: "sigma",
		Password: "Admin@123",
		Email:    "sigma@example.test",
	}
	service.EXPECT().Signup(gomock.Any(), request, time.Minute, time.Hour).
		Return(&models.User{ID: "user-1"}, "access-token", "refresh-token", nil)
	recorder, c := newUserContext()

	(&handler{
		Config: &config.Configuration{Auth: config.ConfigurationAuth{Jwt: config.ConfigurationAuthJwt{
			Ttl:        time.Minute,
			RefreshTTL: time.Hour,
		}}},
		UserSvc: service,
	}).Signup(c, &request)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"token":"access-token"`)
	require.Contains(t, recorder.Body.String(), `"refresh_token":"refresh-token"`)
}

func TestSignupServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := svcuser.NewMockUserService(ctrl)
	request := api.PostUserSignupRequest{
		Username: "sigma",
		Password: "Admin@123",
		Email:    "sigma@example.test",
	}
	service.EXPECT().Signup(gomock.Any(), request, time.Minute, time.Hour).
		Return(nil, "", "", errors.New("failed"))
	recorder, c := newUserContext()

	(&handler{
		Config: &config.Configuration{Auth: config.ConfigurationAuth{Jwt: config.ConfigurationAuthJwt{
			Ttl:        time.Minute,
			RefreshTTL: time.Hour,
		}}},
		UserSvc: service,
	}).Signup(c, &request)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}
