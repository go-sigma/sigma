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

package gc

import (
	"context"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/utils"
)

func (g gc) gcBlobs(ctx context.Context) error {
	blobService := g.blobServiceFactory.New()

	timeTarget := time.Now().Add(-1 * viper.GetDuration("daemon.gc.retention"))

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
			err = g.deleteBlob(ctx, notAssociateBlobs)
			if err != nil {
				return err
			}
		}
		if len(blobs) < pagination {
			break
		}
		curIndex = blobs[len(blobs)-1].ID
	}
	return nil
}

func (g gc) deleteBlob(ctx context.Context, blobs []*models.Blob) error {
	if len(blobs) == 0 {
		return nil
	}
	storageDriver := g.storageDriverFactory.New()
	for _, blob := range blobs {
		err := query.Q.Transaction(func(tx *query.Query) error {
			err := storageDriver.Delete(ctx, utils.GenPathByDigest(digest.Digest(blob.Digest)))
			if err != nil {
				return err
			}
			blobService := g.blobServiceFactory.New(tx)
			err = blobService.DeleteByID(ctx, blob.ID)
			if err != nil {
				return err
			}
			log.Info().Str("digest", blob.Digest).Msg("Delete blob success")
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
