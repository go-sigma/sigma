// Copyright 2023 XImager
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

package manifest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/distribution/distribution/v3/reference"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
)

// DeleteManifest handles the delete manifest request
func (h *handler) DeleteManifest(c echo.Context) error {
	uri := c.Request().URL.Path
	ref := strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/")

	if _, err := digest.Parse(ref); err != nil && !reference.TagRegexp.MatchString(ref) {
		log.Debug().Err(err).Str("ref", ref).Msg("not valid digest or tag")
		return fmt.Errorf("not valid digest or tag")
	}

	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/manifests"), "/v2/")

	ctx := log.Logger.WithContext(c.Request().Context())

	var err error
	var dgest digest.Digest
	if dgest, err = digest.Parse(ref); err == nil { // nolint: staticcheck
		dgest, err = digest.Parse(ref)
		if err != nil {
			log.Error().Err(err).Str("ref", ref).Msg("Parse digest failed")
			return err
		}
		artifactService := h.artifactServiceFactory.New()
		err = artifactService.DeleteByDigest(ctx, repository, dgest.String())
		if err != nil {
			log.Error().Err(err).Str("ref", ref).Msg("Delete artifact failed")
			return err
		}
	} else {
		tagService := h.tagServiceFactory.New()
		err = tagService.DeleteByName(ctx, repository, ref)
		if err != nil {
			return err
		}
	}

	return c.NoContent(http.StatusAccepted)
}
