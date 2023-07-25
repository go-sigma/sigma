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

package logger

import (
	"fmt"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestSetLevel(t *testing.T) {
	SetLevel("debug")
	log.Info().Str("x", "x").Msgf("log level set to %s", zerolog.GlobalLevel())
	SetLevel("info")
	log.Error().Str("x", "x").Err(fmt.Errorf("hello")).Stack().Msgf("log level set to %s", zerolog.GlobalLevel())
}
