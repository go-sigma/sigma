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

package imports

import (
	_ "github.com/go-sigma/sigma/pkg/server/handlers/apidocs"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/artifacts"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/builders"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/caches"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/coderepos"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/daemons"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/namespaces"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/oauth2"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/repositories"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/systems"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/tags"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/tokens"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/users"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/validators"
	_ "github.com/go-sigma/sigma/pkg/server/handlers/webhooks"
)
