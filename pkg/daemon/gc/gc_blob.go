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

package gc

import (
	"context"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func init() {
	workq.TopicHandlers[enums.DaemonGcBlob.String()] = definition.Consumer{
		Handler:     decorator(enums.DaemonGcBlob),
		MaxRetry:    6,
		Concurrency: 10,
		Timeout:     time.Minute * 10,
	}
}

type blobTask struct {
	RunnerID int64
	Blob     models.Blob
}

type gcBlob struct {
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	artifactServiceFactory   dao.ArtifactServiceFactory
	blobServiceFactory       dao.BlobServiceFactory
	daemonServiceFactory     dao.DaemonServiceFactory
	storageDriverFactory     storage.StorageDriverFactory
	config                   configs.Configuration

	deleteBlobChan     chan blobTask
	deleteBlobChanOnce *sync.Once
}

// Run ...
func (g gcBlob) Run(ctx context.Context, runnerID int64, statusChan chan decoratorStatus) error {
	defer close(statusChan)
	statusChan <- decoratorStatus{Daemon: enums.DaemonGcRepository, Status: enums.TaskCommonStatusDoing}

	blobService := g.blobServiceFactory.New()

	timeTarget := time.Now().Add(-1 * g.config.Daemon.Gc.Retention)

	g.deleteBlobChanOnce.Do(g.deleteBlob)

	var curIndex int64
	for {
		blobs, err := blobService.FindWithLastPull(ctx, timeTarget, curIndex, pagination)
		if err != nil {
			return err
		}
		var ids []int64
		for _, blob := range blobs {
			ids = append(ids, blob.ID)
		}
		associateBlobIDs, err := blobService.FindAssociateWithArtifact(ctx, ids)
		if err != nil {
			return err
		}
		notAssociateBlobIDs := mapset.NewSet(ids...)
		notAssociateBlobIDs.RemoveAll(associateBlobIDs...)
		notAssociateBlobSlice := notAssociateBlobIDs.ToSlice()
		if len(notAssociateBlobSlice) > 0 {
			var notAssociateBlobs = make([]*models.Blob, 0, pagination)
			for _, id := range notAssociateBlobSlice {
				for _, blob := range blobs {
					if blob.ID == id {
						notAssociateBlobs = append(notAssociateBlobs, blob)
					}
				}
			}
			if len(notAssociateBlobs) > 0 {
				for _, blob := range notAssociateBlobs {
					g.deleteBlobChan <- blobTask{RunnerID: runnerID, Blob: ptr.To(blob)}
				}
			}
		}
		if len(blobs) < pagination {
			break
		}
		curIndex = blobs[len(blobs)-1].ID
	}
	return nil
}

func (g gcBlob) deleteBlob() {
	ctx := log.Logger.WithContext(context.Background())
	for task := range g.deleteBlobChan {
		err := query.Q.Transaction(func(tx *query.Query) error {
			err := g.blobServiceFactory.New(tx).DeleteByID(ctx, task.Blob.ID)
			if err != nil {
				return err
			}
			err = g.daemonServiceFactory.New(tx).CreateGcBlobRecords(ctx, []*models.DaemonGcBlobRecord{{
				RunnerID: task.RunnerID,
				Digest:   task.Blob.Digest,
			}})
			if err != nil {
				return err
			}
			err = g.storageDriverFactory.New().Delete(ctx, utils.GenPathByDigest(digest.Digest(task.Blob.Digest)))
			if err != nil {
				return err
			}
			log.Info().Str("digest", task.Blob.Digest).Msg("Delete blob success")
			return nil
		})
		if err != nil {
			log.Error().Err(err).Interface("blob", task).Msgf("Delete blob failed: %v", err)
		}
	}
}
