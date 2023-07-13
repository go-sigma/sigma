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
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/daemon"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func init() {
	utils.PanicIf(daemon.RegisterTask(enums.DaemonGc, runner))
}

const pagination = 1000

type gc struct {
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	artifactServiceFactory   dao.ArtifactServiceFactory
	blobServiceFactory       dao.BlobServiceFactory
	storageDriverFactory     storage.StorageDriverFactory
}

func runner(ctx context.Context, task *asynq.Task) error {
	var payload types.DaemonGcPayload
	err := sonic.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return err
	}
	var g = gc{
		namespaceServiceFactory:  dao.NewNamespaceServiceFactory(),
		repositoryServiceFactory: dao.NewRepositoryServiceFactory(),
		artifactServiceFactory:   dao.NewArtifactServiceFactory(),
		blobServiceFactory:       dao.NewBlobServiceFactory(),
		storageDriverFactory:     storage.NewStorageDriverFactory(),
	}
	ctx = log.Logger.WithContext(ctx)
	switch payload.Target {
	case enums.GcTargetBlobsAndArtifacts:
		err = g.gcArtifact(ctx, ptr.To(payload.Scope))
		if err != nil {
			return err
		}
		return g.gcBlobs(ctx)
	case enums.GcTargetArtifacts:
		return g.gcArtifact(ctx, ptr.To(payload.Scope))
	default:
		return fmt.Errorf("payload target is not valid: %s", payload.Target.String())
	}
}
