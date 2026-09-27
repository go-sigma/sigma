// Copyright 2026 sigma
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

package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal/models"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

type initNamespacesParams struct {
	dig.In

	Config              *config.Configuration
	NamespaceRepository reponamespace.NamespaceRepository
}

func initNamespaces(digCon *dig.Container) error {
	return digCon.Invoke(initNamespacesWithParams)
}

// initNamespacesWithParams ensures every namespace declared under
// namespace.initialize exists. It is idempotent and safe to run on every
// startup, so users can declare any number of namespaces with their own
// visibility.
func initNamespacesWithParams(params initNamespacesParams) error {
	ctx := context.Background()
	for _, item := range params.Config.Namespace.Initialize {
		if err := initNamespace(ctx, params.NamespaceRepository, item); err != nil {
			return err
		}
	}
	return nil
}

func initNamespace(ctx context.Context, namespaceRepository reponamespace.NamespaceRepository, item config.ConfigurationNamespaceInit) error {
	_, err := namespaceRepository.GetByName(ctx, item.Name)
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("get namespace failed", "err", err, "namespace", item.Name)
		return fmt.Errorf("get namespace %q failed: %w", item.Name, err)
	}

	err = namespaceRepository.Create(ctx, &models.Namespace{
		ID:         uuid.NewV7String(),
		Name:       item.Name,
		Visibility: item.Visibility,
	})
	if err != nil {
		slog.Error("create namespace failed", "err", err, "namespace", item.Name)
		return fmt.Errorf("create namespace %q failed: %w", item.Name, err)
	}
	slog.Info("namespace initialized", "namespace", item.Name, "visibility", item.Visibility)
	return nil
}
