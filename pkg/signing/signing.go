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
	"context"
	"fmt"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/types/enums"
)

//go:generate mockgen -destination=signing_mocks.go -package=signing github.com/go-sigma/sigma/pkg/signing Signing,Verifying,SigningFactory,VerifyingFactory

// Signing ...
type Signing interface {
	Sign(ctx context.Context, token, priKey, ref string) error
}

// Verifying ...
type Verifying interface {
	Verify(ref, token string) error
}

// Options ...
type Options struct {
	Type enums.SigningType

	Http      bool
	MultiArch bool
}

// SigningFactory is the interface for the signing driver factory
type SigningFactory interface {
	New(http, multiArch bool) (Signing, error)
}

type VerifyingFactory interface {
	New(digCon *dig.Container) (Verifying, error)
}

var verifyingFactories = make(map[enums.SigningType]VerifyingFactory)

var signingFactories = make(map[enums.SigningType]SigningFactory)

// RegisterSigning ...
func RegisterSigning(signingType enums.SigningType, factory SigningFactory) error {
	if _, ok := signingFactories[signingType]; ok {
		return nil
	}
	signingFactories[signingType] = factory
	return nil
}

// RegisterVerifying ...
func RegisterVerifying(signingType enums.SigningType, factory VerifyingFactory) error {
	if _, ok := verifyingFactories[signingType]; ok {
		return nil
	}
	verifyingFactories[signingType] = factory
	return nil
}

// NewSigning ...
func NewSigning(opt Options) (Signing, error) {
	factory, ok := signingFactories[opt.Type]
	if !ok {
		return nil, fmt.Errorf("signing %q not support", opt.Type.String())
	}
	signing, err := factory.New(opt.Http, opt.MultiArch)
	if err != nil {
		return nil, err
	}
	return signing, nil
}

// NewVerifying ...
func NewVerifying(opt Options) (Verifying, error) {
	factory, ok := verifyingFactories[opt.Type]
	if !ok {
		return nil, fmt.Errorf("verifying %q not support", opt.Type.String())
	}
	verifying, err := factory.New(&dig.Container{})
	if err != nil {
		return nil, err
	}
	return verifying, nil
}
