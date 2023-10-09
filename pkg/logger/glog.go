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
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm/logger"
)

// ZLogger is the logger for gorm
type ZLogger struct{}

// LogMode is the log mode
func (l ZLogger) LogMode(logger.LogLevel) logger.Interface {
	return l
}

// Error is the error log
func (l ZLogger) Error(ctx context.Context, msg string, opts ...interface{}) {
	zerolog.Ctx(ctx).Error().Msg(fmt.Sprintf(msg, opts...))
}

// Warn is the warn log
func (l ZLogger) Warn(ctx context.Context, msg string, opts ...interface{}) {
	zerolog.Ctx(ctx).Warn().Msg(fmt.Sprintf(msg, opts...))
}

// Info is the info log
func (l ZLogger) Info(ctx context.Context, msg string, opts ...interface{}) {
	zerolog.Ctx(ctx).Info().Msg(fmt.Sprintf(msg, opts...))
}

// Trace is the trace log
func (l ZLogger) Trace(ctx context.Context, begin time.Time, f func() (string, int64), err error) {
	zl := zerolog.Ctx(ctx)
	var event = zl.Debug()

	event = event.Str("elapsed", time.Since(begin).String())

	sql, rows := f()
	if sql != "" {
		event = event.Str("sql", sql)
	}
	if rows > -1 {
		event = event.Int64("rows", rows)
	}

	event.Send()
}
