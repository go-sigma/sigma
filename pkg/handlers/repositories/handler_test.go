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

package repositories

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	authmocks "github.com/go-sigma/sigma/pkg/auth/mocks"
	daomocks "github.com/go-sigma/sigma/pkg/dal/dao/mocks"
)

func TestFactory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	daoMockAuthServiceFactory := authmocks.NewMockAuthServiceFactory(ctrl)
	daoMockAuditServiceFactory := daomocks.NewMockAuditServiceFactory(ctrl)
	daoMockNamespaceServiceFactory := daomocks.NewMockNamespaceServiceFactory(ctrl)
	daoMockRepositoryServiceFactory := daomocks.NewMockRepositoryServiceFactory(ctrl)
	daoMockArtifactServiceFactory := daomocks.NewMockArtifactServiceFactory(ctrl)
	daoMockTagServiceFactory := daomocks.NewMockTagServiceFactory(ctrl)
	daoMockBuilderServiceFactory := daomocks.NewMockBuilderServiceFactory(ctrl)

	handler := handlerNew(inject{
		authServiceFactory:       daoMockAuthServiceFactory,
		auditServiceFactory:      daoMockAuditServiceFactory,
		namespaceServiceFactory:  daoMockNamespaceServiceFactory,
		repositoryServiceFactory: daoMockRepositoryServiceFactory,
		artifactServiceFactory:   daoMockArtifactServiceFactory,
		tagServiceFactory:        daoMockTagServiceFactory,
		builderServiceFactory:    daoMockBuilderServiceFactory,
	})
	assert.NotNil(t, handler)

	f := factory{}
	err := f.Initialize(echo.New())
	assert.NoError(t, err)
}
