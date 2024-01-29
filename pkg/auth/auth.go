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

//go:generate mockgen -destination=mocks/service.go -package=mocks github.com/go-sigma/sigma/pkg/auth Service
//go:generate mockgen -destination=mocks/service_factory.go -package=mocks github.com/go-sigma/sigma/pkg/auth ServiceFactory

// Service is the interface for the auth service
type Service interface {
	// Namespace ...
	Namespace(c echo.Context, namespaceID int64, auth enums.Auth) (bool, error)
	// Repository ...
	Repository(c echo.Context, repositoryID int64, auth enums.Auth) (bool, error)
	// Tag ...
	Tag(c echo.Context, tagID int64, auth enums.Auth) (bool, error)
	// Artifact ...
	Artifact(c echo.Context, artifactID int64, auth enums.Auth) (bool, error)
}

// ServiceFactory is the interface that provides the artifact service factory methods.
type ServiceFactory interface {
	New() Service
}

type service struct {
	roleServiceFactory       dao.NamespaceMemberServiceFactory
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	tagServiceFactory        dao.TagServiceFactory
	artifactServiceFactory   dao.ArtifactServiceFactory
}

type inject struct {
	roleServiceFactory       dao.NamespaceMemberServiceFactory
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	tagServiceFactory        dao.TagServiceFactory
	artifactServiceFactory   dao.ArtifactServiceFactory
}

type serviceFactory struct {
	roleServiceFactory       dao.NamespaceMemberServiceFactory
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	tagServiceFactory        dao.TagServiceFactory
	artifactServiceFactory   dao.ArtifactServiceFactory
}

// NewServiceFactory creates a new artifact service factory.
func NewServiceFactory(injects ...inject) ServiceFactory {
	roleServiceFactory := dao.NewNamespaceMemberServiceFactory()
	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
	repositoryServiceFactory := dao.NewRepositoryServiceFactory()
	tagServiceFactory := dao.NewTagServiceFactory()
	artifactServiceFactory := dao.NewArtifactServiceFactory()
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
		if ij.tagServiceFactory != nil {
			tagServiceFactory = ij.tagServiceFactory
		}
		if ij.artifactServiceFactory != nil {
			artifactServiceFactory = ij.artifactServiceFactory
		}
	}
	return &serviceFactory{
		roleServiceFactory:       roleServiceFactory,
		namespaceServiceFactory:  namespaceServiceFactory,
		repositoryServiceFactory: repositoryServiceFactory,
		tagServiceFactory:        tagServiceFactory,
		artifactServiceFactory:   artifactServiceFactory,
	}
}

// New ...
func (f *serviceFactory) New() Service {
	s := &service{
		roleServiceFactory:       f.roleServiceFactory,
		namespaceServiceFactory:  f.namespaceServiceFactory,
		repositoryServiceFactory: f.repositoryServiceFactory,
		tagServiceFactory:        f.tagServiceFactory,
		artifactServiceFactory:   f.artifactServiceFactory,
	}
	return s
}

var _ Service = &service{}
