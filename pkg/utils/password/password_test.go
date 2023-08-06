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
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestNew(t *testing.T) {
	pwd := New(20)
	assert.NotNil(t, pwd)
}

func TestHash(t *testing.T) {
	pwdService := New(bcrypt.DefaultCost)
	hashedPwd, err := pwdService.Hash("sigma")
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPwd)

	hashedPwd, err = pwdService.Hash("sigmasigmasigmasigmasigmasigmasigmasigmasigmasigmasigmasigmasigmasigmasigmasigmasigmasigmasigmasigma")
	assert.ErrorIs(t, err, bcrypt.ErrPasswordTooLong)
	assert.Empty(t, hashedPwd)
}

func TestVerify(t *testing.T) {
	pwdService := New()
	hashedPwd, err := pwdService.Hash("sigma")
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPwd)

	eq := pwdService.Verify("sigma", hashedPwd)
	assert.True(t, eq)

	case1 := "invalid"
	neq := pwdService.Verify("sigma", case1)
	assert.False(t, neq)
}
