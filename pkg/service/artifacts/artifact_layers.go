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

package artifacts

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/opencontainers/go-digest"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
)

// manifestRaw is the subset of an image manifest or index the artifact page renders.
type manifestRaw struct {
	MediaType string `json:"mediaType"`
	Layers    []struct {
		MediaType string `json:"mediaType"`
		Size      int64  `json:"size"`
		Digest    string `json:"digest"`
	} `json:"layers"`
}

// imageConfigRaw is the subset of an image config the artifact page renders.
type imageConfigRaw struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	History      []struct {
		Created    string `json:"created"`
		CreatedBy  string `json:"created_by"`
		EmptyLayer bool   `json:"empty_layer"`
	} `json:"history"`
	RootFS struct {
		DiffIDs []string `json:"diff_ids"`
	} `json:"rootfs"`
}

// sbomRaw is the subset of the sbom result the artifact page renders.
type sbomRaw struct {
	Distro struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"distro"`
}

// layerHistory is the image config history entry that produced one layer.
type layerHistory struct {
	command   string
	createdAt string
	diffID    string
}

// GetArtifactLayers gets an artifact of a repository with the layers of its manifest.
func (s *service) GetArtifactLayers(ctx context.Context, namespaceID, repositoryID, digest string) (*api.ArtifactLayers, error) {
	artifactObj, repositoryName, err := s.getArtifactWithReports(ctx, namespaceID, repositoryID, digest)
	if err != nil {
		return nil, err
	}

	manifestBytes, err := s.readManifest(ctx, artifactObj.Digest)
	if err != nil {
		return nil, err
	}

	var manifestObj manifestRaw
	if err = json.Unmarshal(manifestBytes, &manifestObj); err != nil {
		slog.Error("decode artifact manifest failed", "err", err, "artifact_id", artifactObj.ID)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Decode artifact manifest failed: %v", err))
	}

	var configObj imageConfigRaw
	if len(artifactObj.ConfigRaw) > 0 {
		if err = json.Unmarshal(artifactObj.ConfigRaw, &configObj); err != nil {
			// A manifest without a readable config still has layers worth showing.
			slog.Warn("decode artifact config failed", "err", err, "artifact_id", artifactObj.ID)
			configObj = imageConfigRaw{}
		}
	}

	histories := layerHistories(configObj)
	layers := make([]api.ArtifactLayer, 0, len(manifestObj.Layers))
	for idx, item := range manifestObj.Layers {
		layer := api.ArtifactLayer{
			Index:     idx + 1,
			Digest:    item.Digest,
			MediaType: item.MediaType,
			Size:      item.Size,
		}
		if idx < len(histories) {
			layer.Command = histories[idx].command
			layer.CreatedAt = histories[idx].createdAt
			layer.DiffID = histories[idx].diffID
		}
		layers = append(layers, layer)
	}

	var sbomResult sbomRaw
	if len(artifactObj.Sbom.Result) > 0 {
		if err = json.Unmarshal(artifactObj.Sbom.Result, &sbomResult); err != nil {
			slog.Warn("decode artifact sbom failed", "err", err, "artifact_id", artifactObj.ID)
			sbomResult = sbomRaw{}
		}
	}

	return &api.ArtifactLayers{
		ID:           artifactObj.ID,
		Digest:       artifactObj.Digest,
		MediaType:    artifactObj.ContentType,
		Repository:   repositoryName,
		Size:         artifactObj.Size,
		BlobSize:     artifactObj.BlobsSize,
		OS:           configObj.OS,
		Architecture: configObj.Architecture,
		Distro:       distroName(sbomResult),
		PushedAt:     formatArtifactTime(artifactObj.PushedAt),
		PullTimes:    artifactObj.PullTimes,
		Layers:       layers,
	}, nil
}

// getArtifactWithReports resolves the artifact of a repository and validates that the repository
// belongs to the namespace the request names. It also returns the "namespace/name" reference of the
// repository, which the artifact page shows as part of the pull command.
func (s *service) getArtifactWithReports(ctx context.Context, namespaceID, repositoryID, digest string) (*models.Artifact, string, error) {
	namespaceObj, err := s.RepoNs.Get(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Namespace(%s) not found: %v", namespaceID, err))
		}
		return nil, "", errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Namespace(%s) find failed: %v", namespaceID, err))
	}

	repositoryObj, err := s.RepoRegistry.Get(ctx, repositoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Repository(%s) not found: %v", repositoryID, err))
		}
		return nil, "", errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Repository(%s) find failed: %v", repositoryID, err))
	}
	if repositoryObj.NamespaceID != namespaceObj.ID {
		return nil, "", errcode.HTTPErrCodeNotFound.Detail("Repository's namespace ref id not equal namespace id")
	}

	artifactObj, err := s.RepoArtifact.GetByDigestWithReports(ctx, repositoryObj.ID, digest)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Artifact(%s) not found: %v", digest, err))
		}
		return nil, "", errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Artifact(%s) find failed: %v", digest, err))
	}
	return artifactObj, namespaceObj.Name + "/" + repositoryObj.Name, nil
}

// readManifest loads the manifest document of an artifact from storage.
func (s *service) readManifest(ctx context.Context, digestStr string) ([]byte, error) {
	dgst, err := digest.Parse(digestStr)
	if err != nil {
		slog.Error("parse artifact digest failed", "err", err, "digest", digestStr)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Parse artifact digest failed: %v", err))
	}
	reader, err := s.Storage.Reader(ctx, utils.GenManifestPathByDigest(dgst))
	if err != nil {
		slog.Error("read artifact manifest failed", "err", err, "digest", digestStr)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Read artifact manifest failed: %v", err))
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			slog.Warn("close artifact manifest failed", "err", closeErr, "digest", digestStr)
		}
	}()
	manifestBytes, err := io.ReadAll(reader)
	if err != nil {
		slog.Error("read artifact manifest failed", "err", err, "digest", digestStr)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Read artifact manifest failed: %v", err))
	}
	return manifestBytes, nil
}

// layerHistories pairs the non-empty image config history entries with the layer diff ids, which is
// how a layer maps back to the command that created it.
func layerHistories(configObj imageConfigRaw) []layerHistory {
	histories := make([]layerHistory, 0, len(configObj.History))
	for _, item := range configObj.History {
		if item.EmptyLayer {
			continue
		}
		history := layerHistory{
			command:   item.CreatedBy,
			createdAt: formatConfigTime(item.Created),
		}
		if len(histories) < len(configObj.RootFS.DiffIDs) {
			history.diffID = configObj.RootFS.DiffIDs[len(histories)]
		}
		histories = append(histories, history)
	}
	return histories
}

// distroName joins the sbom distro name and version the way the tag list shows it.
func distroName(sbomResult sbomRaw) string {
	if sbomResult.Distro.Name == "" {
		return ""
	}
	if sbomResult.Distro.Version == "" {
		return sbomResult.Distro.Name
	}
	return sbomResult.Distro.Name + " " + sbomResult.Distro.Version
}

// formatArtifactTime renders a millisecond timestamp of the database with the default time pattern.
func formatArtifactTime(ms int64) string {
	if ms == 0 {
		return ""
	}
	return time.UnixMilli(ms).UTC().Format(consts.DefaultTimePattern)
}

// formatConfigTime renders an RFC3339 timestamp of the image config with the default time pattern.
func formatConfigTime(text string) string {
	if text == "" {
		return ""
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		return text
	}
	return parsed.UTC().Format(consts.DefaultTimePattern)
}
