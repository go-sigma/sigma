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
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// SetLevel sets the log level
func SetLevel(levelStr string) {
	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		panic(fmt.Sprintf("invalid log level: %s", levelStr))
	}

	if level < zerolog.TraceLevel || level > zerolog.FatalLevel {
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.DateTime, FormatCaller: func(i any) string {
		var c string
		if cc, ok := i.(string); ok {
			c = cc
		}
		if len(c) > 0 && strings.Contains(c, "/") {
			lastIndex := strings.LastIndex(c, "/")
			left := c[:lastIndex]
			c = c[lastIndex+1:]
			if strings.Contains(left, "/") {
				lastIndex = strings.LastIndex(left, "/")
				c = left[lastIndex+1:] + "/" + c
			}
		}
		return c
	}}).With().Caller().Timestamp().Logger()
}

// Logger is a wrapper for zerolog
type Logger struct{}

// Debug logs a message at Debug level.
func (l *Logger) Debug(args ...any) {
	log.Debug().Msg(fmt.Sprintf("%v", args...))
}

// Info logs a message at Info level.
func (l *Logger) Info(args ...any) {
	log.Info().Msg(fmt.Sprintf("%v", args...))
}

// Warn logs a message at Warning level.
func (l *Logger) Warn(args ...any) {
	log.Warn().Msg(fmt.Sprintf("%v", args...))
}

// Error logs a message at Error level.
func (l *Logger) Error(args ...any) {
	log.Error().Msg(fmt.Sprintf("%v", args...))
}

// Fatal logs a message at Fatal level
// and process will exit with status set to 1.
func (l *Logger) Fatal(args ...any) {
	log.Fatal().Msg(fmt.Sprintf("%v", args...))
}
