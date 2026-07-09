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

package namespace_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

func TestNamespaceMemberRepository(t *testing.T) {
	testkit.InitRepository(t)

	ctx := t.Context()
	namespaceRepository := reponamespace.NewNamespaceRepository()
	memberRepository := reponamespace.NewNamespaceMemberRepository()
	userObj := &models.User{ID: uuid.NewV7String(), Username: "member-user"}
	otherUserObj := &models.User{ID: uuid.NewV7String(), Username: "other-user"}
	namespaceObj := &models.Namespace{ID: uuid.NewV7String(), Name: "member-ns"}
	otherNamespaceObj := &models.Namespace{ID: uuid.NewV7String(), Name: "other-ns"}
	require.NoError(t, query.Q.User.WithContext(ctx).Create(userObj))
	require.NoError(t, query.Q.User.WithContext(ctx).Create(otherUserObj))
	require.NoError(t, namespaceRepository.Create(ctx, namespaceObj))
	require.NoError(t, namespaceRepository.Create(ctx, otherNamespaceObj))

	member, err := memberRepository.AddNamespaceMember(ctx, userObj.ID, *namespaceObj, enums.NamespaceRoleReader)
	require.NoError(t, err)
	otherMember, err := memberRepository.AddNamespaceMember(ctx, userObj.ID, *otherNamespaceObj, enums.NamespaceRoleReader)
	require.NoError(t, err)
	_, err = memberRepository.AddNamespaceMember(ctx, otherUserObj.ID, *namespaceObj, enums.NamespaceRoleManager)
	require.NoError(t, err)

	got, err := memberRepository.GetNamespaceMember(ctx, namespaceObj.ID, userObj.ID)
	require.NoError(t, err)
	require.Equal(t, member.ID, got.ID)
	require.Equal(t, enums.NamespaceRoleReader, got.Role)

	memberships, err := memberRepository.GetNamespacesMember(ctx, []string{namespaceObj.ID, otherNamespaceObj.ID}, userObj.ID)
	require.NoError(t, err)
	require.Len(t, memberships, 2)

	count, err := memberRepository.CountNamespaceMember(ctx, userObj.ID, namespaceObj.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	nameFilter := "member"
	members, total, err := memberRepository.ListNamespaceMembers(ctx, namespaceObj.ID, &nameFilter, testPagination(), testSort("created_at"))
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, members, 1)
	require.Equal(t, userObj.ID, members[0].UserID)
	require.Equal(t, userObj.Username, members[0].User.Username)

	require.NoError(t, memberRepository.UpdateNamespaceMember(ctx, userObj.ID, *namespaceObj, enums.NamespaceRoleManager))
	got, err = memberRepository.GetNamespaceMember(ctx, namespaceObj.ID, userObj.ID)
	require.NoError(t, err)
	require.Equal(t, enums.NamespaceRoleManager, got.Role)

	require.NoError(t, memberRepository.DeleteNamespaceMember(ctx, userObj.ID, *namespaceObj))
	_, err = memberRepository.GetNamespaceMember(ctx, namespaceObj.ID, userObj.ID)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	got, err = memberRepository.GetNamespaceMember(ctx, otherNamespaceObj.ID, userObj.ID)
	require.NoError(t, err)
	require.Equal(t, otherMember.ID, got.ID)
}
