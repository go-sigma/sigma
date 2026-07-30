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

//go:generate mockgen -mock_names Service=MockSystemsService -destination=systems_mocks.go -package=systems github.com/go-sigma/sigma/pkg/service/systems Service

// Service encapsulates system-related business logic.
type Service interface {
	// GetConfig returns the system configuration.
	GetConfig(ctx context.Context) (config.Configuration, error)
	// GetEndpoint returns the HTTP endpoint.
	GetEndpoint(ctx context.Context) (string, error)
	// GetVersion returns the build version info.
	GetVersion(ctx context.Context) (api.GetSystemVersionResponse, error)
}

type service struct {
	dig.In

	Config *config.Configuration
}

func NewService(params service) Service {
	return &params
}

func (s *service) GetConfig(ctx context.Context) (config.Configuration, error) {
	return *s.Config, nil
}

func (s *service) GetEndpoint(ctx context.Context) (string, error) {
	return s.Config.HTTP.Endpoint, nil
}

func (s *service) GetVersion(ctx context.Context) (api.GetSystemVersionResponse, error) {
	return api.GetSystemVersionResponse{
		Version:   version.Version,
		GitHash:   version.GitHash,
		BuildDate: version.BuildDate,
	}, nil
}
