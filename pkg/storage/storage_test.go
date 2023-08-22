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

package storage

import (
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/configs"
)

type dummyFactory struct{}

func (dummyFactory) New(_ configs.Configuration) (StorageDriver, error) {
	return nil, nil
}

type dummyFactoryError struct{}

func (dummyFactoryError) New(_ configs.Configuration) (StorageDriver, error) {
	return nil, fmt.Errorf("dummy error")
}

func TestRegisterDriverFactory(t *testing.T) {
	driverFactories = make(map[string]Factory)

	err := RegisterDriverFactory("dummy", &dummyFactory{})
	assert.NoError(t, err)

	err = RegisterDriverFactory("dummy", &dummyFactory{})
	assert.Error(t, err)
}

func TestInitialize(t *testing.T) {
	driverFactories = make(map[string]Factory)

	err := RegisterDriverFactory("dummy", &dummyFactory{})
	assert.NoError(t, err)

	viper.SetDefault("storage.type", "dummy")
	err = Initialize(configs.Configuration{})
	assert.NoError(t, err)

	viper.SetDefault("storage.type", "fake")
	err = Initialize(configs.Configuration{})
	assert.Error(t, err)

	err = RegisterDriverFactory("dummy-error", &dummyFactoryError{})
	assert.NoError(t, err)

	viper.SetDefault("storage.type", "dummy-error")
	err = Initialize(configs.Configuration{})
	assert.Error(t, err)
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
			want: "test/test",
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
