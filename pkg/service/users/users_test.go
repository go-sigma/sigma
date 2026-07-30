// Copyright 2026 sigma
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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/dal/models"
	repouser "github.com/go-sigma/sigma/pkg/dal/repository/user"
	passwordmocks "github.com/go-sigma/sigma/pkg/service/password/mocks"
	"github.com/go-sigma/sigma/pkg/service/token"
)

func TestListUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepository := repouser.NewMockUserRepository(ctrl)
	items := []*models.User{{ID: "user-1"}}
	userRepository.EXPECT().
		ListWithoutUsername(gomock.Any(), []string{"admin"}, true, nil, api.Pagination{}, api.Sortable{}).
		Return(items, int64(1), nil)

	got, total, err := (&service{RepoUser: userRepository}).ListUsers(
		t.Context(), []string{"admin"}, true, nil, api.Pagination{}, api.Sortable{},
	)
	require.NoError(t, err)
	require.Equal(t, items, got)
	require.Equal(t, int64(1), total)
}

func TestSignupUsesAccessAndRefreshTTL(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepository := repouser.NewMockUserRepository(ctrl)
	passwordSvc := passwordmocks.NewMockService(ctrl)
	tokenSvc := token.NewMockService(ctrl)
	service := &service{
		RepoUser:    userRepository,
		SvcPassword: passwordSvc,
		SvcToken:    tokenSvc,
	}
	request := api.PostUserSignupRequest{
		Username: "sigma",
		Password: "password",
		Email:    "sigma@example.test",
	}
	accessTTL := time.Minute
	refreshTTL := time.Hour

	passwordSvc.EXPECT().Hash(request.Password).Return("hashed", nil)
	userRepository.EXPECT().
		GetByUsername(gomock.Any(), request.Username).
		Return(nil, gorm.ErrRecordNotFound)
	userRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	tokenSvc.EXPECT().New(gomock.Any(), accessTTL).Return("access-token", nil)
	tokenSvc.EXPECT().New(gomock.Any(), refreshTTL).Return("refresh-token", nil)

	user, accessToken, refreshToken, err := service.Signup(
		t.Context(), request, accessTTL, refreshTTL,
	)
	require.NoError(t, err)
	require.Equal(t, request.Username, user.Username)
	require.Equal(t, "access-token", accessToken)
	require.Equal(t, "refresh-token", refreshToken)
}
