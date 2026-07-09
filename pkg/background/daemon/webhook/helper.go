// Copyright 2024 sigma
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

package webhook

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func (w webhook) getNamespace(ctx context.Context, namespaceID *string) (*api.DaemonWebhookNamespace, error) {
	if namespaceID == nil {
		return nil, nil
	}
	namespaceRepository := w.namespaceRepository
	namespaceObj, err := namespaceRepository.Get(ctx, ptr.To(namespaceID))
	if err != nil {
		return nil, err
	}
	repositoryRepository := w.repositoryRepository
	repositoryMapCount, err := repositoryRepository.CountByNamespace(ctx, []string{namespaceObj.ID})
	if err != nil {
		slog.Error("count repository failed", "err", err)
		return nil, err
	}

	tagRepository := w.tagRepository
	tagMapCount, err := tagRepository.CountByNamespace(ctx, []string{namespaceObj.ID})
	if err != nil {
		slog.Error("count tag failed", "err", err)
		return nil, err
	}
	return new(api.DaemonWebhookNamespace{
		ID:              namespaceObj.ID,
		Name:            namespaceObj.Name,
		Description:     namespaceObj.Description,
		Overview:        new(string(namespaceObj.Overview)),
		Visibility:      namespaceObj.Visibility,
		Size:            namespaceObj.Size,
		SizeLimit:       namespaceObj.SizeLimit,
		RepositoryCount: repositoryMapCount[namespaceObj.ID],
		RepositoryLimit: namespaceObj.RepositoryLimit,
		TagCount:        tagMapCount[namespaceObj.ID],
		TagLimit:        namespaceObj.TagLimit,
		CreatedAt:       time.Unix(0, int64(time.Millisecond)*namespaceObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt:       time.Unix(0, int64(time.Millisecond)*namespaceObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
	}), nil
}
