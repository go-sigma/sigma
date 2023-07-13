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

package sbom

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/daemon"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/compress"
)

func init() {
	utils.PanicIf(daemon.RegisterTask(enums.DaemonSbom, daemon.DecoratorArtifact(runner)))
}

func runner(ctx context.Context, artifact *models.Artifact, statusChan chan daemon.DecoratorArtifactStatus) error {
	statusChan <- daemon.DecoratorArtifactStatus{Daemon: enums.DaemonSbom, Status: enums.TaskCommonStatusDoing, Message: ""}
	image := fmt.Sprintf("192.168.31.198:3000/%s@%s", artifact.Repository.Name, artifact.Digest)
	filename := fmt.Sprintf("%s.sbom.json", uuid.New().String())
	cmd := exec.Command("syft", "packages", "-q", "-o", "json", "--file", filename, image)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Error().Err(err).Msg("Run syft failed")
		statusChan <- daemon.DecoratorArtifactStatus{
			Daemon:  enums.DaemonSbom,
			Status:  enums.TaskCommonStatusFailed,
			Stdout:  stdout.Bytes(),
			Stderr:  stderr.Bytes(),
			Message: fmt.Sprintf("Run syft failed: %s", err.Error()),
		}
		return err
	}

	defer func() {
		err := os.Remove(filename)
		if err != nil {
			log.Warn().Err(err).Msg("Remove file failed")
		}
	}()

	compressed, err := compress.Compress(filename)
	if err != nil {
		log.Error().Err(err).Msg("Compress file failed")
		statusChan <- daemon.DecoratorArtifactStatus{Daemon: enums.DaemonSbom, Status: enums.TaskCommonStatusFailed, Message: err.Error()}
		return err
	}

	statusChan <- daemon.DecoratorArtifactStatus{Daemon: enums.DaemonSbom, Status: enums.TaskCommonStatusSuccess, Message: "", Raw: compressed}

	return nil
}
