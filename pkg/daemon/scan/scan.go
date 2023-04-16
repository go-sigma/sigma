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

package scan

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/daemon"
	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/utils/compress"
)

func init() {
	err := daemon.RegisterTask(consts.TopicScan, runner)
	if err != nil {
		log.Fatal().Err(err).Msg("RegisterTask error")
	}
}

func runner(ctx context.Context, atask *asynq.Task) error {
	var task types.TaskScan
	err := json.Unmarshal(atask.Payload(), &task)
	if err != nil {
		log.Error().Err(err).Msg("Unmarshal error")
		return err
	}

	artifactService := dao.NewArtifactService()
	artifact, err := artifactService.Get(ctx, task.ArtifactID)
	if err != nil {
		log.Error().Err(err).Msg("Get artifact failed")
		return err
	}
	image := fmt.Sprintf("127.0.0.1:3000/%s@%s", artifact.Repository.Name, artifact.Digest)

	filename := fmt.Sprintf("%s.trivy.json", uuid.New().String())
	cmd := exec.Command("trivy", "image", "-q", "--format", "json", "--output", filename, "--skip-db-update", "--offline-scan", image)
	out, err := cmd.Output()
	log.Info().Str("out", string(out)).Msg("trivy output")
	if err != nil {
		log.Error().Err(err).Msg("Run trivy failed")
		return err
	}
	defer func() {
		err := os.Remove(filename)
		if err != nil {
			log.Error().Err(err).Msg("Remove file failed")
		}
	}()

	compressed, err := compress.Compress(filename)
	if err != nil {
		log.Error().Err(err).Msg("Compress file failed")
		return err
	}

	log.Info().Int("trivy", len(compressed)).Msg("trivy")
	return nil
}
