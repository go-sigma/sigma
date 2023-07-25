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

package dao

import (
	"context"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

//go:generate mockgen -destination=mocks/auth.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao AuthService
//go:generate mockgen -destination=mocks/auth_factory.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao AuthServiceFactory

// AuthRole defines the role of the user.
type AuthService interface {
	// AddRoleForUser adds a role for a user.
	AddRoleForUser(ctx context.Context, user string, role string, domain string) error
	// DelRoleForUser deletes a role for a user.
	DelRoleForUser(ctx context.Context, user string, role string, domain string) error
}

var _ AuthService = &authService{}

type authService struct {
	tx *query.Query
}

// AuthServiceFactory is the interface that provides the auth service factory methods.
type AuthServiceFactory interface {
	New(txs ...*query.Query) AuthService
}

type authServiceFactory struct{}

// NewAuthServiceFactory creates a new auth service factory.
func NewAuthServiceFactory() AuthServiceFactory {
	return &authServiceFactory{}
}

// New creates a new blob service.
func (f *authServiceFactory) New(txs ...*query.Query) AuthService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &authService{
		tx: tx,
	}
}

// AddRoleForUser adds a role for a user.
func (a *authService) AddRoleForUser(ctx context.Context, user string, role string, domain string) error {
	return a.tx.CasbinRule.WithContext(ctx).Create(&models.CasbinRule{PType: ptr.Of("g"), V0: ptr.Of(user), V1: ptr.Of(role), V2: ptr.Of(domain)})
}

// DelRoleForUser deletes a role for a user.
func (a *authService) DelRoleForUser(ctx context.Context, user string, role string, domain string) error {
	_, err := a.tx.CasbinRule.WithContext(ctx).Where(
		a.tx.CasbinRule.PType.Eq("g"),
		a.tx.CasbinRule.V0.Eq(user),
		a.tx.CasbinRule.V1.Eq(role),
		a.tx.CasbinRule.V2.Eq(domain),
	).Delete()
	return err
}
