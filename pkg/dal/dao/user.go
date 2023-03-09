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

package dao

import (
	"context"

	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
)

// UserService is the interface that provides the user service methods.
type UserService interface {
	// GetByUsername gets the user with the specified user name.
	GetByUsername(ctx context.Context, username string) (*models.User, error)
}

type userService struct {
	tx *query.Query
}

// NewUserService creates a new user service.
func NewUserService(txs ...*query.Query) UserService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &userService{
		tx: tx,
	}
}

// GetByUsername gets the user with the specified user name.
func (s *userService) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return s.tx.User.WithContext(ctx).Where(s.tx.User.Username.Eq(username)).First()
}
