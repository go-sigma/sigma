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

package auth

import (
	"github.com/labstack/echo/v4"

	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

// Service is the interface for the auth service
type Service interface {
	// Namespace ...
	Namespace(c echo.Context, namespaceID int64, auth enums.Auth) bool
	// Repository ...
	Repository(c echo.Context, repositoryID int64, auth enums.Auth) bool
}

var _ Service = &service{}

type service struct {
	roleServiceFactory       dao.NamespaceMemberServiceFactory
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
}

type inject struct {
	roleServiceFactory       dao.NamespaceMemberServiceFactory
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
}

var s Service

// Initialize ...
func Initialize(injects ...inject) {
	roleServiceFactory := dao.NewNamespaceMemberServiceFactory()
	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
	repositoryServiceFactory := dao.NewRepositoryServiceFactory()
	if len(injects) > 0 {
		ij := injects[0]
		if ij.roleServiceFactory != nil {
			roleServiceFactory = ij.roleServiceFactory
		}
		if ij.namespaceServiceFactory != nil {
			namespaceServiceFactory = ij.namespaceServiceFactory
		}
		if ij.repositoryServiceFactory != nil {
			repositoryServiceFactory = ij.repositoryServiceFactory
		}
	}
	s = &service{
		roleServiceFactory:       roleServiceFactory,
		namespaceServiceFactory:  namespaceServiceFactory,
		repositoryServiceFactory: repositoryServiceFactory,
	}
}

// GetInstance ...
func GetInstance() Service {
	return s
}