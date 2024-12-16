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

package scan

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	syftTypes "github.com/anchore/syft/syft/format/syftjson/model"
	"github.com/anchore/syft/syft/source"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/compress"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/token"
)

func init() {
	workq.TopicHandlers[enums.DaemonSbom] = definition.Consumer{
		Handler:     decorator(runnerSbom),
		MaxRetry:    1,
		Concurrency: 10,
		Timeout:     time.Minute * 10,
	}
}

// reportDistro ...
type reportDistro struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// reportSbom ...
type reportSbom struct {
	Distro       reportDistro `json:"distro"`
	Os           string       `json:"os"`
	Architecture string       `json:"architecture"`
}

func runnerSbom(ctx context.Context, artifact *models.Artifact, statusChan chan decoratorArtifactStatus) error {
	defer close(statusChan)
	statusChan <- decoratorArtifactStatus{Daemon: enums.DaemonSbom, Status: enums.TaskCommonStatusDoing, Message: ""}

	config := ptr.To(configs.GetConfiguration())
	userService := dao.NewUserServiceFactory().New()
	userObj, err := userService.GetByUsername(ctx, consts.UserInternal)
	if err != nil {
		return err
	}
	tokenService, err := token.New(dig.New()) // TODO: dig
	if err != nil {
		return err
	}
	authorization, err := tokenService.New(userObj.ID, config.Auth.Jwt.Ttl)
	if err != nil {
		return err
	}

	image := fmt.Sprintf("%s/%s@%s", utils.TrimHTTP(config.HTTP.InternalEndpoint), artifact.Repository.Name, artifact.Digest)
	filename := fmt.Sprintf("%s.sbom.json", uuid.New().String())

	cmd := exec.Command("syft", "scan", "-q", "-o", fmt.Sprintf("json=%s", filename), fmt.Sprintf("registry:%s", image)) // nolint: gosec
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = append(cmd.Env, fmt.Sprintf("SYFT_REGISTRY_AUTH_TOKEN=%s", authorization))
	if strings.HasPrefix(config.HTTP.InternalEndpoint, "http://") {
		cmd.Env = append(cmd.Env, "SYFT_REGISTRY_INSECURE_USE_HTTP=true")
	}
	if strings.HasPrefix(config.HTTP.InternalEndpoint, "https://") {
		cmd.Env = append(cmd.Env, "SYFT_REGISTRY_INSECURE_SKIP_TLS_VERIFY=true")
	}

	log.Info().Str("artifactDigest", artifact.Digest).Strs("env", cmd.Env).Str("cmd", cmd.String()).Msg("Start sbom artifact")

	defer func() {
		if utils.IsFile(filename) {
			err := os.Remove(filename)
			if err != nil {
				log.Warn().Err(err).Msg("Remove file failed")
			}
		}
	}()

	err = cmd.Run()
	if err != nil {
		var stdoutBytes = stdout.Bytes()
		var stderrBytes = stderr.Bytes()
		log.Error().Err(err).Str("stdout", string(stdoutBytes)).Str("stderr", string(stderrBytes)).Msg("Run syft failed")
		statusChan <- decoratorArtifactStatus{
			Daemon:  enums.DaemonSbom,
			Status:  enums.TaskCommonStatusFailed,
			Stdout:  stdoutBytes,
			Stderr:  stderrBytes,
			Message: fmt.Sprintf("Run syft failed: %s", err.Error()),
		}
		return err
	}

	var syftObj syftTypes.Document
	fileContent, err := os.Open(filename)
	if err != nil {
		log.Error().Err(err).Str("filename", filename).Msg("Open sbom file failed")
		statusChan <- decoratorArtifactStatus{
			Daemon:  enums.DaemonSbom,
			Status:  enums.TaskCommonStatusFailed,
			Stdout:  []byte(""),
			Stderr:  []byte(""),
			Message: fmt.Sprintf("Open sbom file(%s) failed: %v", filename, err),
		}
		return err
	}
	defer func() {
		fileContent.Close() // nolint: errcheck
	}()
	err = json.NewDecoder(fileContent).Decode(&syftObj)
	if err != nil {
		log.Error().Err(err).Str("filename", filename).Msg("Decode sbom file failed")
		statusChan <- decoratorArtifactStatus{
			Daemon:  enums.DaemonSbom,
			Status:  enums.TaskCommonStatusFailed,
			Stdout:  []byte(""),
			Stderr:  []byte(""),
			Message: fmt.Sprintf("Decode sbom file(%s) failed: %v", filename, err),
		}
		return err
	}
	var report = reportSbom{
		Distro: reportDistro{
			Name:    syftObj.Distro.ID,
			Version: syftObj.Distro.VersionID,
		},
	}
	syftMetadata, ok := syftObj.Source.Metadata.(source.ImageMetadata)
	if ok {
		report.Os = syftMetadata.OS
		report.Architecture = syftMetadata.Architecture
	}
	reportBytes, err := json.Marshal(report)
	if err != nil {
		log.Error().Err(err).Msg("Marshal report failed")
		statusChan <- decoratorArtifactStatus{Daemon: enums.DaemonVulnerability, Status: enums.TaskCommonStatusFailed, Message: err.Error()}
		return err
	}

	compressed, err := compress.Compress(filename)
	if err != nil {
		log.Error().Err(err).Msg("Compress file failed")
		statusChan <- decoratorArtifactStatus{Daemon: enums.DaemonSbom, Status: enums.TaskCommonStatusFailed, Message: err.Error()}
		return err
	}

	log.Info().Str("artifactDigest", artifact.Digest).Msg("Success sbom artifact")

	statusChan <- decoratorArtifactStatus{Daemon: enums.DaemonSbom,
		Status:  enums.TaskCommonStatusSuccess,
		Message: "", Raw: compressed, Result: reportBytes}

	return nil
}
