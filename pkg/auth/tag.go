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
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

// Tag ...
func (s authService) Tag(user models.User, tagID int64, auth enums.Auth) (bool, error) {
	ctx := log.Logger.WithContext(context.Background())

	tagService := s.tagServiceFactory.New()
	tagObj, err := tagService.GetByID(ctx, tagID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("tagID", tagID).Msg("Get tag by id failed")
			return false, errors.Join(err, fmt.Errorf("Get tag by id(%d) failed", tagID))
		}
		log.Error().Err(err).Int64("tagID", tagID).Msg("Get tag by id not found")
		return false, errors.Join(err, fmt.Errorf("Get tag by id(%d) not found", tagID))
	}
	return s.Repository(user, tagObj.RepositoryID, auth)
}
