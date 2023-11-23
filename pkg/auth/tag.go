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
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/types/enums"
)

// Tag ...
func (s service) Tag(c echo.Context, tagID int64, auth enums.Auth) bool {
	ctx := log.Logger.WithContext(c.Request().Context())

	tagService := s.tagServiceFactory.New()
	tagObj, err := tagService.GetByID(ctx, tagID)
	if err != nil {
		log.Error().Err(err).Msg("Get tag by id failed")
		return false
	}
	return s.Repository(c, tagObj.RepositoryID, auth)
}
