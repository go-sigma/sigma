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

package locker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-sigma/sigma/pkg/configs"
)

// Lock ...
type Lock interface {
	// Unlock ...
	Unlock() error
}

// Locker ...
type Locker interface {
	// Lock ...
	Lock(ctx context.Context, name string, expire time.Duration) (Lock, error)
}

// LockerFactory ...
type LockerFactory interface {
	// New ...
	New(config configs.Configuration) (Locker, error)
}

// LockerFactories ...
var LockerFactories = make(map[string]LockerFactory, 5)

// LockerClient ...
var LockerClient Locker

// Initialize ...
func Initialize(config configs.Configuration) error {
	l, ok := LockerFactories[strings.ToLower(config.Locker.Type.String())]
	if !ok {
		return fmt.Errorf("Locker %s not support", strings.ToLower(config.Locker.Type.String()))
	}
	var err error
	LockerClient, err = l.New(config)
	if err != nil {
		return err
	}
	return nil
}
