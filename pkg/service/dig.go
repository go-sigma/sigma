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
		digCon.Provide(namespaces.NewService),
		digCon.Provide(analytics.NewService),
		digCon.Provide(repositories.NewService),
		digCon.Provide(tags.NewService),
		digCon.Provide(artifacts.NewService),
		digCon.Provide(users.NewService),
		digCon.Provide(webhooks.NewService),
		digCon.Provide(builders.NewService),
		digCon.Provide(coderepos.NewService),
		digCon.Provide(daemons.NewService),
		digCon.Provide(systems.NewService),
		digCon.Provide(oauth2.NewService),
		digCon.Provide(blob.NewService),
		digCon.Provide(manifest.NewService),
		digCon.Provide(upload.NewService),
	} {
		if err := e; err != nil {
			return err
		}
	}
	return nil
}
