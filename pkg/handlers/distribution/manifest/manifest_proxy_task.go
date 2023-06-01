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
	"github.com/distribution/distribution/v3"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/dal/models"
)

func proxyArtifactTask(c echo.Context, repository, digest, contentType string, manifestBytes []byte) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	proxyServiceFactory := dao.NewProxyServiceFactory()
	proxyService := proxyServiceFactory.New()

	manifest, _, err := distribution.UnmarshalManifest(contentType, manifestBytes)
	if err != nil {
		return err
	}

	var proxyArtifactBlobs = make([]models.ProxyArtifactTaskBlob, 0, len(manifest.References()))
	for _, desc := range manifest.References() {
		proxyArtifactBlobs = append(proxyArtifactBlobs, models.ProxyArtifactTaskBlob{Blob: desc.Digest.String()})
	}

	err = proxyService.SaveProxyArtifact(ctx, &models.ProxyArtifactTask{
		Repository:  repository,
		Digest:      digest,
		Size:        uint64(len(manifestBytes)),
		ContentType: contentType,
		Blobs:       proxyArtifactBlobs,
		Raw:         manifestBytes,
	})
	if err != nil {
		return nil
	}

	return nil
}
