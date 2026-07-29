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
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	svcuser "github.com/go-sigma/sigma/pkg/service/users"
)

func TestLogout(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := svcuser.NewMockUserService(ctrl)
	service.EXPECT().Logout(gomock.Any(), []string{"access-token", "refresh-token"}, "jti-1").Return(nil)
	recorder, c := newUserContext()
	c.Set(consts.ContextJti, "jti-1")

	(&handler{UserSvc: service}).Logout(c, &api.PostUserLogoutRequest{
		Tokens: []string{"access-token", "refresh-token"},
	})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestLogoutWithoutJTI(t *testing.T) {
	recorder, c := newUserContext()

	(&handler{}).Logout(c, &api.PostUserLogoutRequest{Tokens: []string{"access-token"}})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}
