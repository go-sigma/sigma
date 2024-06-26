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

package definition

import "context"

//go:generate mockgen -destination=mocks/signing.go -package=mocks github.com/go-sigma/sigma/pkg/signing/definition Signing
//go:generate mockgen -destination=mocks/verifying.go -package=mocks github.com/go-sigma/sigma/pkg/signing/definition Verifying

// Signing ...
type Signing interface {
	Sign(ctx context.Context, token, priKey, ref string) error
}

// Verifying ...
type Verifying interface {
	Verify(ref, token string) error
}
