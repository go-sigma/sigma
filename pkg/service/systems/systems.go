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

package systems

import (
	"context"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/version"
)

//go:generate mockgen -destination=systems_mocks.go -package=systems github.com/go-sigma/sigma/pkg/service/systems SystemsService

// SystemsService encapsulates system-related business logic.
type SystemsService interface {
	// GetConfig returns the system configuration.
	GetConfig(ctx context.Context) (config.Configuration, error)
	// GetEndpoint returns the HTTP endpoint.
	GetEndpoint(ctx context.Context) (string, error)
	// GetVersion returns the build version info.
	GetVersion(ctx context.Context) (api.GetSystemVersionResponse, error)
}

type systemsService struct {
	config *config.Configuration
}

type ServiceParams struct {
	dig.In

	Config *config.Configuration
}

func NewService(digCon *dig.Container) error {
	return digCon.Provide(func(params ServiceParams) SystemsService {
		return &systemsService{config: params.Config}
	})
}

func (s *systemsService) GetConfig(ctx context.Context) (config.Configuration, error) {
	return *s.config, nil
}

func (s *systemsService) GetEndpoint(ctx context.Context) (string, error) {
	return s.config.HTTP.Endpoint, nil
}

func (s *systemsService) GetVersion(ctx context.Context) (api.GetSystemVersionResponse, error) {
	return api.GetSystemVersionResponse{
		Version:   version.Version,
		GitHash:   version.GitHash,
		BuildDate: version.BuildDate,
	}, nil
}
