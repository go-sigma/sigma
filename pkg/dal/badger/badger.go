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

package badger

import (
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/configs"
)

type logger struct{}

// Errorf is the error log
func (l logger) Errorf(msg string, opts ...interface{}) {
	log.Error().Msg(strings.TrimSpace(fmt.Sprintf(msg, opts...)))
}

// Warningf is the warning log
func (l logger) Warningf(msg string, opts ...interface{}) {
	log.Warn().Msg(strings.TrimSpace(fmt.Sprintf(msg, opts...)))
}

// Infof is the info log
func (l logger) Infof(msg string, opts ...interface{}) {
	log.Info().Msg(strings.TrimSpace(fmt.Sprintf(msg, opts...)))
}

// Debugf is the debug log
func (l logger) Debugf(msg string, opts ...interface{}) {
	log.Debug().Msg(strings.TrimSpace(fmt.Sprintf(msg, opts...)))
}

// New new badger instance
func New(config configs.Configuration) (*badger.DB, error) {
	client, err := badger.Open(badger.DefaultOptions(config.Badger.Path).WithLogger(&logger{}).
		WithLoggingLevel(badger.DEBUG))
	if err != nil {
		return nil, err
	}
	return client, nil
}
