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

package daemon

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"

	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/types/enums"
)

// DecoratorArtifactStatus is a status for decorator
type DecoratorArtifactStatus struct {
	Daemon  enums.Daemon
	Status  enums.TaskCommonStatus
	Raw     []byte
	Stdout  []byte
	Stderr  []byte
	Message string
}

// DecoratorArtifact is a decorator for daemon task runners
func DecoratorArtifact(runner func(context.Context, *models.Artifact, chan DecoratorArtifactStatus) error) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, atask *asynq.Task) error {
		log.Info().Msg("got a task")

		artifactServiceFactory := dao.NewArtifactServiceFactory()
		artifactService := artifactServiceFactory.New()

		id := gjson.GetBytes(atask.Payload(), "artifact_id").Int()
		artifact, err := artifactService.Get(ctx, id)
		if err != nil {
			return err
		}

		var statusChan = make(chan DecoratorArtifactStatus, 1)
		defer close(statusChan)
		go func() {
			for status := range statusChan {
				switch status.Daemon {
				case enums.DaemonVulnerability:
					err = artifactService.SaveVulnerability(context.Background(), &models.ArtifactVulnerability{
						ArtifactID: id,
						Raw:        status.Raw,
						Status:     status.Status,
						Stdout:     status.Stdout,
						Stderr:     status.Stderr,
						Message:    status.Message,
					})
				case enums.DaemonSbom:
					err = artifactService.SaveSbom(context.Background(), &models.ArtifactSbom{
						ArtifactID: id,
						Raw:        status.Raw,
						Status:     status.Status,
						Stdout:     status.Stdout,
						Stderr:     status.Stderr,
						Message:    status.Message,
					})
				default:
					continue
				}
				if err != nil {
					log.Error().Err(err).Msg("Update artifact status failed")
				}
			}
		}()

		err = runner(ctx, artifact, statusChan)
		if err != nil {
			return err
		}

		return nil
	}
}
