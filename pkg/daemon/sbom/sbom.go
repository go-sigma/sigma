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

package sbom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	syftTypes "github.com/anchore/syft/syft/formats/syftjson/model"
	"github.com/anchore/syft/syft/source"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/go-sigma/sigma/pkg/daemon"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/compress"
)

func init() {
	workq.TopicHandlers[enums.DaemonSbom.String()] = definition.Consumer{
		Handler:     daemon.DecoratorArtifact(runner),
		MaxRetry:    6,
		Concurrency: 10,
		Timeout:     time.Minute * 10,
	}
}

// ReportDistro ...
type ReportDistro struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Report ...
type Report struct {
	Distro       ReportDistro `json:"distro"`
	Os           string       `json:"os"`
	Architecture string       `json:"architecture"`
}

func runner(ctx context.Context, artifact *models.Artifact, statusChan chan daemon.DecoratorArtifactStatus) error {
	statusChan <- daemon.DecoratorArtifactStatus{Daemon: enums.DaemonSbom, Status: enums.TaskCommonStatusDoing, Message: ""}
	image := fmt.Sprintf("%s/%s@%s", utils.TrimHTTP(viper.GetString("server.internalEndpoint")), artifact.Repository.Name, artifact.Digest)
	filename := fmt.Sprintf("%s.sbom.json", uuid.New().String())
	cmd := exec.Command("syft", "packages", "-q", "-o", "json", "--file", filename, image)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Info().Str("artifactDigest", artifact.Digest).Str("command", cmd.String()).Msg("Start sbom artifact")

	defer func() {
		err := os.Remove(filename)
		if err != nil {
			log.Warn().Err(err).Msg("Remove file failed")
		}
	}()

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

	var syftObj syftTypes.Document
	fileContent, err := os.Open(filename)
	if err != nil {
		log.Error().Err(err).Str("filename", filename).Msg("Open sbom file failed")
		statusChan <- daemon.DecoratorArtifactStatus{
			Daemon:  enums.DaemonSbom,
			Status:  enums.TaskCommonStatusFailed,
			Stdout:  []byte(""),
			Stderr:  []byte(""),
			Message: fmt.Sprintf("Open sbom file(%s) failed: %v", filename, err),
		}
		return err
	}
	err = json.NewDecoder(fileContent).Decode(&syftObj)
	if err != nil {
		log.Error().Err(err).Str("filename", filename).Msg("Decode sbom file failed")
		statusChan <- daemon.DecoratorArtifactStatus{
			Daemon:  enums.DaemonSbom,
			Status:  enums.TaskCommonStatusFailed,
			Stdout:  []byte(""),
			Stderr:  []byte(""),
			Message: fmt.Sprintf("Decode sbom file(%s) failed: %v", filename, err),
		}
		return err
	}
	var report = Report{
		Distro: ReportDistro{
			Name:    syftObj.Distro.ID,
			Version: syftObj.Distro.VersionID,
		},
	}
	syftMetadata, ok := syftObj.Source.Metadata.(source.StereoscopeImageSourceMetadata)
	if ok {
		report.Os = syftMetadata.OS
		report.Architecture = syftMetadata.Architecture
	}
	reportBytes, err := json.Marshal(report)
	if err != nil {
		log.Error().Err(err).Msg("Marshal report failed")
		statusChan <- daemon.DecoratorArtifactStatus{Daemon: enums.DaemonVulnerability, Status: enums.TaskCommonStatusFailed, Message: err.Error()}
		return err
	}

	compressed, err := compress.Compress(filename)
	if err != nil {
		log.Error().Err(err).Msg("Compress file failed")
		statusChan <- daemon.DecoratorArtifactStatus{Daemon: enums.DaemonSbom, Status: enums.TaskCommonStatusFailed, Message: err.Error()}
		return err
	}

	log.Info().Str("artifactDigest", artifact.Digest).Msg("Success sbom artifact")

	statusChan <- daemon.DecoratorArtifactStatus{Daemon: enums.DaemonSbom, Status: enums.TaskCommonStatusSuccess, Message: "", Raw: compressed, Result: reportBytes}

	return nil
}
