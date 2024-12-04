// Copyright 2024 sigma
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

package tests

import (
	"os"
	"strings"

	"github.com/spf13/viper"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

// GetConfig gets the configuration from the environment variables
func GetConfig() (*configs.Configuration, error) {
	viper.AutomaticEnv()
	viper.SetEnvPrefix(consts.AppName)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := viper.Unmarshal(configs.GetConfiguration())
	if err != nil {
		return nil, err
	}
	config := configs.GetConfiguration()
	badgerDir, err := os.MkdirTemp("", "badger")
	if err != nil {
		return nil, err
	}
	config.Badger.Path = badgerDir
	config.Badger.Enabled = true
	config.Locker.Type = enums.LockerTypeBadger

	return configs.GetConfiguration(), nil
}
