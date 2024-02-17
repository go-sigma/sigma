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
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func (w webhook) getNamespace(ctx context.Context, namespaceID *int64) (*types.DaemonWebhookNamespace, error) {
	if namespaceID == nil {
		return nil, nil
	}
	namespaceService := w.namespaceServiceFactory.New()
	namespaceObj, err := namespaceService.Get(ctx, ptr.To(namespaceID))
	if err != nil {
		return nil, err
	}
	repositoryService := w.repositoryServiceFactory.New()
	repositoryMapCount, err := repositoryService.CountByNamespace(ctx, []int64{namespaceObj.ID})
	if err != nil {
		log.Error().Err(err).Msg("Count repository failed")
		return nil, err
	}

	tagService := w.tagServiceFactory.New()
	tagMapCount, err := tagService.CountByNamespace(ctx, []int64{namespaceObj.ID})
	if err != nil {
		log.Error().Err(err).Msg("Count tag failed")
		return nil, err
	}
	return ptr.Of(types.DaemonWebhookNamespace{
		ID:              namespaceObj.ID,
		Name:            namespaceObj.Name,
		Description:     namespaceObj.Description,
		Overview:        ptr.Of(string(namespaceObj.Overview)),
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
