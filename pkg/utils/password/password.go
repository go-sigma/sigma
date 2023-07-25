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

package password

import (
	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -destination=mocks/password.go -package=mocks github.com/ximager/ximager/pkg/utils/password Password

// Password is an interface for password hashing
// nolint: revive
type Password interface {
	// Hash returns a hashed password
	Hash(pwd string) (string, error)
	// Verify compares a password with a hashed password
	Verify(pwd, hash string) bool
}

type password struct {
	cost int
}

const (
	// DefaultCost is the default cost for bcrypt
	DefaultCost = 10
)

// New returns a new password instance
func New(costs ...int) Password {
	var cost = DefaultCost
	if len(costs) > 0 {
		cost = costs[0]
	}
	return &password{
		cost: cost,
	}
}

// Hash returns a hashed password
func (p *password) Hash(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), p.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Verify compares a password with a hashed password
func (p *password) Verify(pwd, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd)) == nil
}
