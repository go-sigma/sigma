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
	"errors"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/types/enums"
)

// Repository ...
func (s service) Repository(c echo.Context, repositoryID int64, auth enums.Auth) (bool, error) {
	ctx := log.Logger.WithContext(c.Request().Context())
	repositoryService := s.repositoryServiceFactory.New()
	repositoryObj, err := repositoryService.Get(ctx, repositoryID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("repositoryID", repositoryID).Msg("Get repository by id failed")
			return false, errors.Join(err, fmt.Errorf("Get repository by id(%d) failed", repositoryID))
		}
		log.Error().Err(err).Int64("repositoryID", repositoryID).Msg("Get repository by id not found")
		return false, errors.Join(err, fmt.Errorf("Get repository by id(%d) not found", repositoryID))
	}
	return s.Namespace(c, repositoryObj.NamespaceID, auth)
}
