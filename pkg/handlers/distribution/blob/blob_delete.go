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

package blob

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/xerrors"
)

// DeleteBlob handles the delete blob request
// Note: if blob associate with artifact, it cannot be deleted
func (h *handler) DeleteBlob(c echo.Context) error {
	uri := c.Request().URL.Path

	dgest, err := digest.Parse(strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/"))
	if err != nil {
		log.Error().Err(err).Str("digest", c.QueryParam("digest")).Msg("Parse digest failed")
		return err
	}
	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/blobs"), "/v2/")
	log.Debug().Str("digest", dgest.String()).Str("repository", repository).Msg("Blob info")

	ctx := log.Logger.WithContext(c.Request().Context())

	blobService := h.blobServiceFactory.New()
	blobObj, err := blobService.FindByDigest(ctx, dgest.String())
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Str("digest", dgest.String()).Msg("Find blob failed")
			return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
		}
		log.Error().Err(err).Str("digest", dgest.String()).Msg("Parse content length failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeBlobUnknown)
	}
	result, err := blobService.FindAssociateWithArtifact(ctx, []int64{blobObj.ID})
	if err != nil {
		log.Error().Err(err).Str("digest", dgest.String()).Msg("Find associate with artifact failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}
	if len(result) > 0 {
		log.Error().Err(err).Str("digest", dgest.String()).Msg("Blob associate with artifact")
		return xerrors.NewDSError(c, xerrors.DSErrCodeBlobAssociated)
	}

	err = blobService.DeleteByID(ctx, blobObj.ID)
	if err != nil {
		log.Error().Err(err).Str("digest", dgest.String()).Msg("Delete blob failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}

	return c.NoContent(http.StatusAccepted)
}
