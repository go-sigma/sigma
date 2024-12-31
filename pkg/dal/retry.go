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

package dal

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/dal/query"
)

// TxnWithRetry ...
func TxnWithRetry(fc func(tx *query.Query) error, opts ...*sql.TxOptions) error {
	randInt, err := rand.Int(rand.Reader, big.NewInt(300))
	if err != nil {
		return fmt.Errorf("failed to generate random number: %w", err)
	}
	err = retry.Do(func() error {
		return query.Q.Transaction(fc, opts...)
	}, retry.MaxDelay(time.Second*10), retry.Attempts(6), retry.LastErrorOnly(true),
		retry.Delay(300*time.Millisecond+time.Duration(randInt.Int64())*time.Millisecond),
		retry.RetryIf(func(err error) bool {
			if err != nil && strings.Contains(err.Error(), "Deadlock") {
				log.Debug().Err(err).Msg("transaction deadlock, retry again now")
				return true
			}
			return false
		}))
	if err != nil {
		return err
	}
	return nil
}
