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

package webhooks

import (
	"errors"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// checkWebhookAuth checks if the user has permission to access the webhook based on namespace
// authLevel specifies the required authorization level (e.g., AuthRead, AuthManage)
// Returns an error if authorization fails, nil if authorized
func (h *handler) checkWebhookAuth(c echo.Context, user *models.User, webhook *models.Webhook, authLevel enums.Auth) error {
	if webhook.NamespaceID == nil {
		// Global webhook - requires admin or root role
		if !(user.Role == enums.UserRoleAdmin || user.Role == enums.UserRoleRoot) { // nolint: staticcheck
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "No permission with this api")
		}
	} else {
		// Namespace webhook - check namespace permission
		namespaceID := ptr.To(webhook.NamespaceID)
		authChecked, err := h.AuthServiceFactory.New().Namespace(ptr.To(user), namespaceID, authLevel)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error().Err(err).Int64("NamespaceID", namespaceID).Msg("Namespace not found")
				return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, fmt.Sprintf("Namespace(%d) not found: %v", namespaceID, err))
			}
			log.Error().Err(err).Int64("NamespaceID", namespaceID).Msg("Namespace find failed")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Namespace(%d) find failed: %v", namespaceID, err))
		}
		if !authChecked {
			log.Error().Int64("UserID", user.ID).Int64("NamespaceID", namespaceID).Msg("Auth check failed")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "No permission with this api")
		}
	}
	return nil
}
