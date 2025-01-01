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

	"github.com/go-sigma/sigma/pkg/dal/dao"
)

func initDigContainer(digCon *dig.Container) error {
	for _, e := range []error{
		digCon.Provide(dao.NewArtifactServiceFactory),
		digCon.Provide(dao.NewAuditServiceFactory),
		digCon.Provide(dao.NewAuthRuleServiceFactory),
		digCon.Provide(dao.NewAuthRoleServiceFactory),
		digCon.Provide(dao.NewBlobServiceFactory),
		digCon.Provide(dao.NewBlobUploadServiceFactory),
		digCon.Provide(dao.NewBuilderServiceFactory),
		digCon.Provide(dao.NewCodeRepositoryServiceFactory),
		digCon.Provide(dao.NewDaemonServiceFactory),
		digCon.Provide(dao.NewNamespaceMemberServiceFactory),
		digCon.Provide(dao.NewNamespaceServiceFactory),
		digCon.Provide(dao.NewRepositoryServiceFactory),
		digCon.Provide(dao.NewSettingServiceFactory),
		digCon.Provide(dao.NewTagServiceFactory),
		digCon.Provide(dao.NewUserServiceFactory),
		digCon.Provide(dao.NewWebhookServiceFactory),
		digCon.Provide(dao.NewWorkQueueServiceFactory),
	} {
		if err := e; err != nil {
			return err
		}
	}
	return nil
}
