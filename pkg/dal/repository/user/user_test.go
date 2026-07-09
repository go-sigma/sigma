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

package user_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repouser "github.com/go-sigma/sigma/pkg/dal/repository/user"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

func TestNewUserRepository(t *testing.T) {
	require.NotNil(t, repouser.NewUserRepository())
	require.NotNil(t, repouser.NewUserRepository(query.Q))
}

func TestUserRepository(t *testing.T) {
	testkit.InitRepository(t)

	ctx := t.Context()
	userRepository := repouser.NewUserRepository()
	password := "test-case"
	email := "email"
	userObj := &models.User{
		ID:       uuid.NewV7String(),
		Username: "test-case",
		Password: &password,
		Email:    &email,
		Role:     enums.UserRoleUser,
	}
	adminObj := &models.User{
		ID:       uuid.NewV7String(),
		Username: "admin-case",
		Role:     enums.UserRoleAdmin,
	}
	require.NoError(t, userRepository.Create(ctx, userObj))
	require.NoError(t, userRepository.Create(ctx, adminObj))

	testUser, err := userRepository.Get(ctx, userObj.ID)
	require.NoError(t, err)
	require.Equal(t, userObj.Username, testUser.Username)

	testUser, err = userRepository.GetByUsername(ctx, "test-case")
	require.NoError(t, err)
	require.Equal(t, ptr.To(testUser.Password), "test-case")

	total, err := userRepository.Count(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)

	nameFilter := "test"
	users, total, err := userRepository.List(ctx, &nameFilter, testPagination(), testSort("username", enums.SortMethodAsc))
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, users, 1)
	require.Equal(t, userObj.ID, users[0].ID)

	users, total, err = userRepository.ListWithoutUsername(ctx, []string{"nobody"}, true, nil, testPagination(), testSort("username", enums.SortMethodAsc))
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, users, 1)
	require.Equal(t, userObj.ID, users[0].ID)

	require.NoError(t, userRepository.UpdateByID(ctx, userObj.ID, map[string]any{
		query.User.NamespaceLimit.ColumnName().String(): int64(3),
	}))
	require.NoError(t, userRepository.UpdateByID(ctx, userObj.ID, map[string]any{}))
	testUser, err = userRepository.Get(ctx, userObj.ID)
	require.NoError(t, err)
	require.Equal(t, int64(3), testUser.NamespaceLimit)

	err = userRepository.UpdateByID(ctx, uuid.NewV7String(), map[string]any{
		query.User.NamespaceLimit.ColumnName().String(): int64(1),
	})
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestUserRepositoryThirdParty(t *testing.T) {
	testkit.InitRepository(t)

	ctx := t.Context()
	userRepository := repouser.NewUserRepository()
	userObj := &models.User{ID: uuid.NewV7String(), Username: "third-party"}
	accountID := "account-1"
	token := "token"
	updatedToken := "updated-token"
	user3rdPartyObj := &models.User3rdParty{
		ID:        uuid.NewV7String(),
		UserID:    userObj.ID,
		Provider:  enums.ProviderGithub,
		AccountID: &accountID,
		Token:     &token,
	}
	require.NoError(t, userRepository.Create(ctx, userObj))
	require.NoError(t, userRepository.CreateUser3rdParty(ctx, user3rdPartyObj))

	got, err := userRepository.GetUser3rdPartyByAccountID(ctx, enums.ProviderGithub, accountID)
	require.NoError(t, err)
	require.Equal(t, userObj.ID, got.User.ID)

	got, err = userRepository.GetUser3rdPartyByProvider(ctx, userObj.ID, enums.ProviderGithub)
	require.NoError(t, err)
	require.Equal(t, user3rdPartyObj.ID, got.ID)
	require.Equal(t, userObj.ID, got.User.ID)

	got, err = userRepository.GetUser3rdParty(ctx, user3rdPartyObj.ID)
	require.NoError(t, err)
	require.Equal(t, user3rdPartyObj.ID, got.ID)

	thirdParties, err := userRepository.ListUser3rdParty(ctx, userObj.ID)
	require.NoError(t, err)
	require.Len(t, thirdParties, 1)

	require.NoError(t, userRepository.UpdateUser3rdParty(ctx, user3rdPartyObj.ID, map[string]any{
		query.User3rdParty.Token.ColumnName().String(): updatedToken,
	}))
	require.NoError(t, userRepository.UpdateUser3rdParty(ctx, user3rdPartyObj.ID, map[string]any{}))
	got, err = userRepository.GetUser3rdParty(ctx, user3rdPartyObj.ID)
	require.NoError(t, err)
	require.Equal(t, updatedToken, ptr.To(got.Token))

	err = userRepository.UpdateUser3rdParty(ctx, uuid.NewV7String(), map[string]any{
		query.User3rdParty.Token.ColumnName().String(): "missing",
	})
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestUserRepositoryRecoverCode(t *testing.T) {
	testkit.InitRepository(t)

	ctx := t.Context()
	userRepository := repouser.NewUserRepository()
	userObj := &models.User{ID: uuid.NewV7String(), Username: "recover"}
	recoverCode := &models.UserRecoverCode{
		ID:     uuid.NewV7String(),
		UserID: userObj.ID,
		Code:   "recover-code",
	}
	require.NoError(t, userRepository.Create(ctx, userObj))
	require.NoError(t, userRepository.CreateRecoverCode(ctx, recoverCode))

	gotRecoverCode, err := userRepository.GetRecoverCodeByUserID(ctx, userObj.ID)
	require.NoError(t, err)
	require.Equal(t, recoverCode.ID, gotRecoverCode.ID)

	gotUser, err := userRepository.GetByRecoverCode(ctx, recoverCode.Code)
	require.NoError(t, err)
	require.Equal(t, userObj.ID, gotUser.ID)

	require.NoError(t, userRepository.DeleteRecoverCode(ctx, userObj.ID))
	_, err = userRepository.GetRecoverCodeByUserID(ctx, userObj.ID)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func testPagination() api.Pagination {
	page := 1
	limit := 10
	return api.Pagination{Page: &page, Limit: &limit}
}

func testSort(column string, method enums.SortMethod) api.Sortable {
	return api.Sortable{Sort: &column, Method: &method}
}
