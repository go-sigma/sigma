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
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/daemon"
	"github.com/ximager/ximager/pkg/types"
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
	filename := fmt.Sprintf("%s.trivy.json", uuid.New().String())
	cmd := exec.Command("trivy", "image", "-q", "--format", "json", "--output", filename, "--skip-db-update", "--offline-scan", task.Image)
	out, err := cmd.Output()
	log.Info().Str("out", string(out)).Msg("trivy output")
	if err != nil {
		log.Error().Err(err).Msg("Run trivy failed")
		return err
	}
	file, err := os.Open(filename)
	if err != nil {
		log.Error().Err(err).Msg("Open file failed")
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Error().Err(err).Msg("Close file failed")
		}
		err = os.Remove(filename)
		if err != nil {
			log.Error().Err(err).Msg("Remove file failed")
		}
	}()
	var trivy bytes.Buffer
	gzipWriter, err := gzip.NewWriterLevel(&trivy, gzip.BestSpeed)
	if err != nil {
		log.Error().Err(err).Msg("Create gzip reader failed")
		return err
	}
	_, err = io.Copy(gzipWriter, file)
	if err != nil {
		log.Error().Err(err).Msg("Copy file to gzip reader failed")
		return err
	}
	err = gzipWriter.Close()
	if err != nil {
		log.Error().Err(err).Msg("Close gzip reader failed")
	}
	log.Info().Int("trivy", len(trivy.Bytes())).Msg("trivy")
	return nil
}
