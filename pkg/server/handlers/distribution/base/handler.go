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

package distribution

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/server/handlers/distribution"
	svcrepository "github.com/go-sigma/sigma/pkg/service/repositories"
	svctag "github.com/go-sigma/sigma/pkg/service/tags"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler is the interface for the distribution handlers
type Handler interface {
	// GetHealthy handles the get healthy request
	GetHealthy(ctx *gin.Context)
	// ListTags handles the list tags request
	ListTags(ctx *gin.Context)
	// ListRepositories handles the list repositories request
	ListRepositories(ctx *gin.Context)
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	Config     *config.Configuration
	RepoSvc    svcrepository.RepositoryService
	TagSvc     svctag.TagService
	Authorizer authz.Authorizer
}

type factory struct{}

// Initialize initializes the distribution manifest handlers
func (f factory) Initialize(c *gin.Context, digCon *dig.Container) error {
	method := c.Request.Method
	uri := c.Request.RequestURI
	return digCon.Invoke(func(h handler) error {
		if method == http.MethodGet {
			switch {
			case uri == "/v2/":
				h.GetHealthy(c)
				return nil
			case uri == "/v2/_catalog":
				h.ListRepositories(c)
				return nil
			case strings.HasSuffix(uri, "/tags/list") && strings.HasPrefix(uri, "/v2/"):
				h.ListTags(c)
				return nil
			}
		}
		return distribution.ErrNext
	})
}

func init() {
	utils.PanicIf(distribution.RegisterRouterFactory(&factory{}, 1))
}
