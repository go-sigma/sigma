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

package builders

import (
	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/utils"
)

// actingUserID returns the id of the authenticated user, or an empty string when the request has
// no session user. Builds triggered without a user fall back to the builder creator.
func actingUserID(c *gin.Context) string {
	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok || user == nil {
		return ""
	}
	return user.ID
}
