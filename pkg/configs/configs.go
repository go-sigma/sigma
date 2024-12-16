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

package configs

import "github.com/go-sigma/sigma/pkg/utils/ptr"

type checker func(cfg Configuration) error

var checkers []checker

// Initialize initializes the configs.
func Initialize() error {
	defaultSettings()
	return nil
}

// CheckMiddleware ...
func CheckMiddleware() error {
	for _, checker := range checkers {
		err := checker(ptr.To(configuration))
		if err != nil {
			return err
		}
	}
	return nil
}
