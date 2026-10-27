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

package credential

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	reposetting "github.com/go-sigma/sigma/pkg/dal/repository/setting"
	"github.com/go-sigma/sigma/pkg/testkit"
)

// fakeSettingRepository is an in-memory SettingRepository.
type fakeSettingRepository struct {
	values map[string][]byte
}

func (f *fakeSettingRepository) Create(_ context.Context, key string, val []byte) error {
	f.values[key] = val
	return nil
}

func (f *fakeSettingRepository) Update(_ context.Context, key string, val []byte) error {
	f.values[key] = val
	return nil
}

func (f *fakeSettingRepository) Delete(_ context.Context, key string) error {
	delete(f.values, key)
	return nil
}

func (f *fakeSettingRepository) Get(_ context.Context, key string) (*models.Setting, error) {
	val, ok := f.values[key]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &models.Setting{Key: key, Val: val}, nil
}

var _ reposetting.SettingRepository = &fakeSettingRepository{}

func TestNewFromConfig(t *testing.T) {
	key := strings.Repeat("k", MinMasterKeyLength)
	svc, err := New(Params{
		Config:            &config.Configuration{Auth: config.ConfigurationAuth{CredentialKey: key}},
		SettingRepository: &fakeSettingRepository{values: map[string][]byte{}},
	})
	require.NoError(t, err)

	sealed, err := svc.Seal("ghp_sigma")
	require.NoError(t, err)
	require.NotEqual(t, "ghp_sigma", sealed)

	plaintext, err := svc.Open(sealed)
	require.NoError(t, err)
	require.Equal(t, "ghp_sigma", plaintext)
}

func TestNewFromSettings(t *testing.T) {
	key := strings.Repeat("s", MinMasterKeyLength)
	svc, err := New(Params{
		Config: &config.Configuration{},
		SettingRepository: &fakeSettingRepository{values: map[string][]byte{
			consts.SettingCredentialMasterKey: []byte(key),
		}},
	})
	require.NoError(t, err)

	sealed, err := svc.Seal("token")
	require.NoError(t, err)
	plaintext, err := svc.Open(sealed)
	require.NoError(t, err)
	require.Equal(t, "token", plaintext)
}

func TestNewRejectsShortConfigKey(t *testing.T) {
	_, err := New(Params{
		Config:            &config.Configuration{Auth: config.ConfigurationAuth{CredentialKey: "too-short"}},
		SettingRepository: &fakeSettingRepository{values: map[string][]byte{}},
	})
	require.Error(t, err)
}

func TestNewWithoutKey(t *testing.T) {
	testkit.InitRepository(t)

	_, err := New(Params{
		Config:            &config.Configuration{},
		SettingRepository: reposetting.NewSettingRepository(),
	})
	require.Error(t, err)
}

func TestOpenWithAnotherKey(t *testing.T) {
	first, err := New(Params{
		Config:            &config.Configuration{Auth: config.ConfigurationAuth{CredentialKey: strings.Repeat("a", MinMasterKeyLength)}},
		SettingRepository: &fakeSettingRepository{values: map[string][]byte{}},
	})
	require.NoError(t, err)
	sealed, err := first.Seal("ghp_sigma")
	require.NoError(t, err)

	second, err := New(Params{
		Config:            &config.Configuration{Auth: config.ConfigurationAuth{CredentialKey: strings.Repeat("b", MinMasterKeyLength)}},
		SettingRepository: &fakeSettingRepository{values: map[string][]byte{}},
	})
	require.NoError(t, err)
	_, err = second.Open(sealed)
	require.Error(t, err)
}
