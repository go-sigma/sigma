// Copyright 2024 sigma
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

package dal

import (
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/authz"
	repoanalytics "github.com/go-sigma/sigma/pkg/dal/repository/analytics"
	repoaudit "github.com/go-sigma/sigma/pkg/dal/repository/audit"
	repobuilder "github.com/go-sigma/sigma/pkg/dal/repository/builder"
	repocoderepo "github.com/go-sigma/sigma/pkg/dal/repository/coderepo"
	repodaemon "github.com/go-sigma/sigma/pkg/dal/repository/daemon"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	reposetting "github.com/go-sigma/sigma/pkg/dal/repository/setting"
	repouser "github.com/go-sigma/sigma/pkg/dal/repository/user"
	repowebhook "github.com/go-sigma/sigma/pkg/dal/repository/webhook"
)

func initDigContainer(digCon *dig.Container) error {
	for _, e := range []error{
		digCon.Provide(reporegistry.NewArtifactRepository),
		digCon.Provide(repoanalytics.NewAnalyticsRepository),
		digCon.Provide(repoaudit.NewAuditRepository),
		digCon.Provide(reporegistry.NewBlobRepository),
		digCon.Provide(repobuilder.NewBuilderRepository),
		digCon.Provide(repocoderepo.NewCodeRepositoryRepository),
		digCon.Provide(repodaemon.NewDaemonRepository),
		digCon.Provide(reponamespace.NewNamespaceMemberRepository),
		digCon.Provide(reponamespace.NewCachedNamespaceRepository),
		digCon.Provide(reporegistry.NewCachedRepositoryRepository),
		digCon.Provide(reposetting.NewSettingRepository),
		digCon.Provide(reporegistry.NewTagRepository),
		digCon.Provide(repouser.NewUserRepository),
		digCon.Provide(repowebhook.NewWebhookRepository),
		digCon.Provide(authz.NewAuthorizer),
	} {
		if err := e; err != nil {
			return err
		}
	}
	return nil
}
