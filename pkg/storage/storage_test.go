// Copyright 2025 sigma
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

package storage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
)

type dummyFactory struct{}

func (dummyFactory) New(_ *config.Configuration) (StorageDriver, error) {
	return nil, nil
}

type dummyFactoryError struct{}

func (dummyFactoryError) New(_ *config.Configuration) (StorageDriver, error) {
	return nil, fmt.Errorf("dummy error")
}

func TestRegister(t *testing.T) {
	factories = make(map[enums.StorageType]Factory)

	err := Register(enums.StorageTypeDummy, &dummyFactory{})
	assert.NoError(t, err)

	err = Register(enums.StorageTypeDummy, &dummyFactory{})
	assert.Error(t, err)
}

func TestInitialize(t *testing.T) {
	factories = make(map[enums.StorageType]Factory)

	err := Register(enums.StorageTypeDummy, &dummyFactory{})
	assert.NoError(t, err)

	_, err = Initialize(&config.Configuration{
		Storage: config.ConfigurationStorage{
			Type: enums.StorageTypeDummy,
		},
	})
	assert.NoError(t, err)

	_, err = Initialize(&config.Configuration{
		Storage: config.ConfigurationStorage{
			Type: "fake",
		},
	})
	assert.Error(t, err)

	err = Register("dummy-error", &dummyFactoryError{})
	assert.NoError(t, err)
}

func TestSanitizePath(t *testing.T) {
	type args struct {
		rootDirectory string
		p             string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test-1",
			args: args{
				rootDirectory: "",
				p:             "test",
			},
			want: "test",
		},
		{
			name: "test-2",
			args: args{
				rootDirectory: ".",
				p:             "test",
			},
			want: "test",
		},
		{
			name: "test-3",
			args: args{
				rootDirectory: "./",
				p:             "test",
			},
			want: "test",
		},
		{
			name: "test-4",
			args: args{
				rootDirectory: "/",
				p:             "test",
			},
			want: "test",
		},
		{
			name: "test-5",
			args: args{
				rootDirectory: "/test",
				p:             "test",
			},
			want: "/test/test",
		},
		{
			name: "test-6",
			args: args{
				rootDirectory: "./test",
				p:             "test",
			},
			want: "test/test",
		},
		{
			name: "test-7",
			args: args{
				rootDirectory: "test",
				p:             "test",
			},
			want: "test/test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizePath(tt.args.rootDirectory, tt.args.p); got != tt.want {
				t.Errorf("SanitizePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
