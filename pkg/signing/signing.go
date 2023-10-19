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

package signing

import (
	"github.com/go-sigma/sigma/pkg/signing/cosign/sign"
	"github.com/go-sigma/sigma/pkg/signing/definition"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

// Options ...
type Options struct {
	Type enums.SigningType

	Http      bool
	Multiarch bool
}

// NewSigning ...
func NewSigning(opt Options) definition.Signing {
	switch opt.Type {
	case enums.SigningTypeCosign:
		return sign.New(opt.Http, opt.Multiarch)
	default:
		return sign.New(opt.Http, opt.Multiarch)
	}
}
