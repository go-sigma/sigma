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
	"bytes"
	"fmt"

	"github.com/distribution/distribution/v3"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/utils/hash"
)

func (h *handler) proxyTaskArtifact(c echo.Context, repository, digest, contentType string, manifestBytes []byte) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	proxyTaskService := h.proxyTaskServiceFactory.New()

	manifest, _, err := distribution.UnmarshalManifest(contentType, manifestBytes)
	if err != nil {
		return err
	}

	var proxyArtifactBlobs = make([]models.ProxyTaskArtifactBlob, 0, len(manifest.References()))
	for _, desc := range manifest.References() {
		proxyArtifactBlobs = append(proxyArtifactBlobs, models.ProxyTaskArtifactBlob{Blob: desc.Digest.String()})
	}

	err = proxyTaskService.SaveProxyTaskArtifact(ctx, &models.ProxyTaskArtifact{
		Repository:  repository,
		Digest:      digest,
		Size:        uint64(len(manifestBytes)),
		ContentType: contentType,
		Blobs:       proxyArtifactBlobs,
		Raw:         manifestBytes,
	})
	if err != nil {
		return err
	}

	return nil
}

func (h *handler) proxyTaskTag(c echo.Context, repository, reference, contentType string, manifestBytes []byte) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	proxyTaskService := h.proxyTaskServiceFactory.New()

	manifest, _, err := distribution.UnmarshalManifest(contentType, manifestBytes)
	if err != nil {
		return err
	}

	var proxyTaskTagManifest []models.ProxyTaskTagManifest
	if contentType == "application/vnd.docker.distribution.manifest.v2+json" || contentType == "application/vnd.oci.image.manifest.v1+json" {
		h, err := hash.Reader(bytes.NewReader(manifestBytes))
		if err != nil {
			return err
		}
		proxyTaskTagManifest = []models.ProxyTaskTagManifest{{Digest: fmt.Sprintf("sha256:%s", h)}}
	} else if contentType == "application/vnd.docker.distribution.manifest.list.v2+json" || contentType == "application/vnd.oci.image.index.v1+json" {
		proxyTaskTagManifest = make([]models.ProxyTaskTagManifest, 0, len(manifest.References()))
		for _, desc := range manifest.References() {
			proxyTaskTagManifest = append(proxyTaskTagManifest, models.ProxyTaskTagManifest{Digest: desc.Digest.String()})
		}
	}

	err = proxyTaskService.SaveProxyTaskTag(ctx, &models.ProxyTaskTag{
		Repository:  repository,
		Reference:   reference,
		Size:        uint64(len(manifestBytes)),
		ContentType: contentType,
		Raw:         manifestBytes,
		Manifests:   proxyTaskTagManifest,
	})
	if err != nil {
		return err
	}

	return nil
}
