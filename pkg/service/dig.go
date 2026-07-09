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

package service

import (
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/service/analytics"
	"github.com/go-sigma/sigma/pkg/service/artifacts"
	"github.com/go-sigma/sigma/pkg/service/builders"
	"github.com/go-sigma/sigma/pkg/service/coderepos"
	"github.com/go-sigma/sigma/pkg/service/daemons"
	"github.com/go-sigma/sigma/pkg/service/distribution/blob"
	"github.com/go-sigma/sigma/pkg/service/distribution/manifest"
	"github.com/go-sigma/sigma/pkg/service/distribution/upload"
	"github.com/go-sigma/sigma/pkg/service/namespaces"
	"github.com/go-sigma/sigma/pkg/service/oauth2"
	"github.com/go-sigma/sigma/pkg/service/repositories"
	"github.com/go-sigma/sigma/pkg/service/systems"
	"github.com/go-sigma/sigma/pkg/service/tags"
	"github.com/go-sigma/sigma/pkg/service/users"
	"github.com/go-sigma/sigma/pkg/service/webhooks"
)

// InitDigContainer registers all services into the dig container.
func InitDigContainer(digCon *dig.Container) error {
	for _, e := range []error{
		namespaces.NewService(digCon),
		analytics.NewService(digCon),
		repositories.NewService(digCon),
		tags.NewService(digCon),
		artifacts.NewService(digCon),
		users.NewService(digCon),
		webhooks.NewService(digCon),
		builders.NewService(digCon),
		coderepos.NewService(digCon),
		daemons.NewService(digCon),
		systems.NewService(digCon),
		oauth2.NewService(digCon),
		blob.NewService(digCon),
		manifest.NewService(digCon),
		upload.NewService(digCon),
	} {
		if err := e; err != nil {
			return err
		}
	}
	return nil
}
