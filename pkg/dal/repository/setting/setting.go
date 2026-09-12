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

package setting

import (
	"context"

	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

//go:generate go tool mockgen -destination=setting_mocks.go -package=setting github.com/go-sigma/sigma/pkg/dal/repository/setting SettingRepository

// SettingRepository defines setting repository operations
type SettingRepository interface {
	// Create creates a new setting
	Create(ctx context.Context, key string, val []byte) error
	// Update updates a setting by key
	Update(ctx context.Context, key string, val []byte) error
	// Delete deletes a setting by key
	Delete(ctx context.Context, key string) error
	// Get gets a setting by key
	Get(ctx context.Context, key string) (*models.Setting, error)
}

type settingRepository struct {
	tx *query.Query
}

// NewSettingRepository creates a new setting repository with the optional query transaction
func NewSettingRepository(txs ...*query.Query) SettingRepository {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &settingRepository{
		tx: tx,
	}
}

// Create creates a new setting
func (s settingRepository) Create(ctx context.Context, key string, val []byte) error {
	var setting = models.Setting{ID: uuid.NewV7String(), Key: key, Val: val}
	return s.tx.Setting.WithContext(ctx).Create(&setting)
}

// Update updates a setting by key
func (s settingRepository) Update(ctx context.Context, key string, val []byte) error {
	var setting = models.Setting{Key: key, Val: val}
	_, err := s.tx.Setting.WithContext(ctx).Where(s.tx.Setting.Key.Eq(key)).Updates(&setting)
	return err
}

// Delete deletes a setting by key
func (s settingRepository) Delete(ctx context.Context, key string) error {
	matched, err := s.tx.Setting.WithContext(ctx).Unscoped().Where(s.tx.Setting.Key.Eq(key)).Delete()
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Get gets a setting by key
func (s settingRepository) Get(ctx context.Context, key string) (*models.Setting, error) {
	return s.tx.Setting.WithContext(ctx).Where(s.tx.Setting.Key.Eq(key)).First()
}
