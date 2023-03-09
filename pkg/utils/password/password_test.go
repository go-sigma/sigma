// Copyright 2023 XImager
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
)

func TestNew(t *testing.T) {
	var param = Params{}
	pwd := New(param)
	assert.NotNil(t, pwd)
}

func TestHash(t *testing.T) {
	pwdService := New()
	hashedPwd, err := pwdService.Hash("ximager")
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPwd)
}

func TestVerify(t *testing.T) {
	pwdService := New()
	hashedPwd, err := pwdService.Hash("ximager")
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPwd)

	eq, err := pwdService.Verify("ximager", hashedPwd)
	assert.NoError(t, err)
	assert.True(t, eq)

	case1 := "$argon2id1$v=19$m=65536,t=1,p=2$eIH0HZBTb0P1bu3mOw1xyQ$hUCbRhWG0ouJIW9+gFWDMu/w727820HkRA6bxpkRA5w"
	_, err = pwdService.Verify("ximager", case1)
	assert.Equal(t, ErrIncompatibleVariant, err)

	case2 := "$argon2id$v=18$m=65536,t=1,p=2$eIH0HZBTb0P1bu3mOw1xyQ$hUCbRhWG0ouJIW9+gFWDMu/w727820HkRA6bxpkRA5w"
	_, err = pwdService.Verify("ximager", case2)
	assert.Equal(t, ErrIncompatibleVersion, err)

	case3 := "$argo$n2id1$v=19$m=65536,t=1,p=2$eIH0HZBTb0P1bu3mOw1xyQ$hUCbRhWG0ouJIW9+gFWDMu/w727820HkRA6bxpkRA5w"
	_, err = pwdService.Verify("ximager", case3)
	assert.Equal(t, ErrInvalidHash, err)

	case4 := "$argon2id$v=19$m=65536,t=1,p=2$x7fdU5ghyVkaXmCL5Yt9Pg$ylFKFx4QnVqUXqQ73gjqBAL424FvjfCClVCBur9/SSM"
	eq, err = pwdService.Verify("ximager", case4)
	assert.NoError(t, err)
	assert.False(t, eq)

	case5 := "$argon2id$v=19$m=65536,t=1,p=2$x7fdU5ghyVkaXmCL5Yt9Pg$ylFKFx4QnVqUXqQ73gjqBAL424FvjfCClVCBur9/SSwq="
	_, err = pwdService.Verify("ximager", case5)
	assert.Error(t, err)

	case6 := "$argon2id$v=19$m=65536,t=1,p=2$x7fdU5ghyVkaXmCL5Yt9Pg+12$ylFKFx4QnVqUXqQ73gjqBAL424FvjfCClVCBur9/SSwq="
	_, err = pwdService.Verify("ximager", case6)
	assert.Error(t, err)
}
