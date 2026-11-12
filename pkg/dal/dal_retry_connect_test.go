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

package dal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/config"
)

func TestRetryConnectRetriesUnreachableDatabase(t *testing.T) {
	var attempts int
	err := retryConnect(t.Context(), func() error {
		attempts++
		if attempts < 3 {
			return &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}
		}
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 3, attempts)
}

func TestRetryConnectFailsFastOnConfigurationError(t *testing.T) {
	var attempts int
	err := retryConnect(t.Context(), func() error {
		attempts++
		return fmt.Errorf("unknown database type: bogus")
	})
	require.Error(t, err)
	require.Equal(t, 1, attempts)
}

// TestRetryConnectStopsWhenBudgetExpires pins the contract
// connectDatabaseWithRetry relies on: the budget expires as
// context.DeadlineExceeded, not as the last attempt error.
func TestRetryConnectStopsWhenBudgetExpires(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
	defer cancel()

	var attempts int
	err := retryConnect(ctx, func() error {
		attempts++
		return &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}
	})
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Greater(t, attempts, 0)
}

func TestConnectDatabaseWithRetryDoesNotBlameBudgetForConfigError(t *testing.T) {
	db, err := connectDatabaseWithRetry(&config.Configuration{})
	require.Nil(t, db)
	require.ErrorContains(t, err, "unknown database type")
	require.NotContains(t, err.Error(), "unreachable after")
}

func TestIsRetryableConnectError(t *testing.T) {
	refused := &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "postgres connect error wraps a dial error",
			err:  fmt.Errorf("failed to connect: %w", errors.Join(refused, io.EOF)),
			want: true,
		},
		{
			name: "mysql returns the dial error directly",
			err:  refused,
			want: true,
		},
		{
			name: "dns lookup failure",
			err:  &net.DNSError{Err: "no such host"},
			want: true,
		},
		{
			name: "postgres is still recovering",
			err:  fmt.Errorf("connect: %w", &pgconn.PgError{Code: pgerrDatabaseStartingUp}),
			want: true,
		},
		{
			name: "bad postgres credentials",
			err:  &pgconn.PgError{Code: "28P01"},
			want: false,
		},
		{
			name: "bad mysql credentials",
			err:  &mysqldriver.MySQLError{Number: 1045, Message: "access denied"},
			want: false,
		},
		{
			name: "unknown database type",
			err:  fmt.Errorf("unknown database type: bogus"),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, isRetryableConnectError(tt.err))
		})
	}
}
