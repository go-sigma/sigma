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

package artifact

import (
	"context"
	"database/sql"
	"time"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"

	"github.com/ximager/ximager/pkg/daemon"
	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
	"github.com/ximager/ximager/pkg/types/enums"
	"github.com/ximager/ximager/pkg/utils"
)

func init() {
	utils.PanicIf(daemon.RegisterTask(enums.DaemonProxyArtifact, newRunner()))
}

// when a new blob is pulled bypass the proxy or pushed a new blob to the registry, the proxy will be notified

type inject struct {
	proxyServiceFactory      dao.ProxyServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	artifactServiceFactory   dao.ArtifactServiceFactory
	blobServiceFactory       dao.BlobServiceFactory
}

func newRunner(injects ...inject) func(ctx context.Context, atask *asynq.Task) error {
	repositoryServiceFactory := dao.NewRepositoryServiceFactory()
	artifactServiceFactory := dao.NewArtifactServiceFactory()
	blobServiceFactory := dao.NewBlobServiceFactory()
	proxyServiceFactory := dao.NewProxyServiceFactory()
	if len(injects) > 0 {
		ij := injects[0]
		if ij.proxyServiceFactory != nil {
			proxyServiceFactory = ij.proxyServiceFactory
		}
		if ij.repositoryServiceFactory != nil {
			repositoryServiceFactory = ij.repositoryServiceFactory
		}
		if ij.artifactServiceFactory != nil {
			artifactServiceFactory = ij.artifactServiceFactory
		}
		if ij.blobServiceFactory != nil {
			blobServiceFactory = ij.blobServiceFactory
		}
	}
	return func(ctx context.Context, atask *asynq.Task) error {
		proxyService := proxyServiceFactory.New()
		blobID := gjson.GetBytes(atask.Payload(), "blob_digest").String()
		artifactTasks, err := proxyService.FindByBlob(ctx, blobID)
		if err != nil {
			return err
		}

		if len(artifactTasks) == 0 {
			log.Debug().Str("blob", blobID).Msg("Cannot find any task for this blob")
			return nil
		}

		blobService := blobServiceFactory.New()
		for _, task := range artifactTasks {
			var blobDigests = make([]string, 0, len(task.Blobs))
			for _, blob := range task.Blobs {
				blobDigests = append(blobDigests, blob.Blob)
			}
			var allExist = true
			blobs, err := blobService.FindByDigests(ctx, blobDigests)
			if err != nil {
				log.Error().Err(err).Strs("digests", blobDigests).Msg("Find digests failed")
				return err
			} else if len(blobs) != len(task.Blobs) {
				log.Debug().Strs("digests", blobDigests).Msg("Not all digests are exist")
				allExist = false
			}
			if allExist {
				err := query.Q.Transaction(func(tx *query.Query) error {
					repositoryService := repositoryServiceFactory.New(tx)
					repository := &models.Repository{Name: task.Repository}
					err := repositoryService.Save(ctx, repository)
					if err != nil {
						return err
					}
					artifactService := artifactServiceFactory.New(tx)
					artifact := &models.Artifact{
						RepositoryID: repository.ID,
						Digest:       task.Digest,
						Size:         task.Size,
						ContentType:  task.ContentType,
						Raw:          string(task.Raw),
						PushedAt:     time.Now(),
						PullTimes:    0,
						LastPull:     sql.NullTime{},

						Blobs: blobs,
					}
					err = artifactService.Save(ctx, artifact)
					if err != nil {
						return err
					}
					return nil
				})
				if err != nil {
					log.Error().Err(err).Msg("Create artifact failed")
					return err
				}
				log.Info().Str("repository", task.Repository).Str("artifact", task.Digest).Msg("Proxy artifact task success")
			}
		}
		return nil
	}
}
