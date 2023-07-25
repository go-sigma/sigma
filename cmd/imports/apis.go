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

package cmd

import (
	_ "github.com/go-sigma/sigma/pkg/handlers/apidocs"
	_ "github.com/go-sigma/sigma/pkg/handlers/artifacts"
	_ "github.com/go-sigma/sigma/pkg/handlers/namespaces"
	_ "github.com/go-sigma/sigma/pkg/handlers/oauth2"
	_ "github.com/go-sigma/sigma/pkg/handlers/repositories"
	_ "github.com/go-sigma/sigma/pkg/handlers/systems"
	_ "github.com/go-sigma/sigma/pkg/handlers/tags"
	_ "github.com/go-sigma/sigma/pkg/handlers/tokens"
	_ "github.com/go-sigma/sigma/pkg/handlers/users"
	_ "github.com/go-sigma/sigma/pkg/handlers/validators"
)
