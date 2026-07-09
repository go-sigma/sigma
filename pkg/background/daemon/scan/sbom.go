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
	"log/slog"
	"os"
	"os/exec"
	"strings"

	syftTypes "github.com/anchore/syft/syft/format/syftjson/model"
	"github.com/anchore/syft/syft/source"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/compress"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

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

func runnerSbom(ctx context.Context, p params, artifact *models.Artifact, statusChan chan decoratorArtifactStatus) error {
	defer close(statusChan)
	statusChan <- decoratorArtifactStatus{Daemon: enums.DaemonSbom, Status: enums.TaskCommonStatusDoing, Message: ""}

	userRepository := p.UserRepository
	userObj, err := userRepository.GetByUsername(ctx, consts.UserInternal)
	if err != nil {
		return err
	}
	authorization, err := p.TokenSvc.New(userObj.ID, p.Config.Auth.Jwt.Ttl)
	if err != nil {
		return err
	}

	image := fmt.Sprintf("%s/%s@%s", utils.TrimHTTP(p.Config.HTTP.InternalEndpoint), artifact.Repository.Name, artifact.Digest)
	filename := fmt.Sprintf("%s.sbom.json", uuid.NewV7String())

	cmd := exec.Command("syft", "scan", "-q", "-o", fmt.Sprintf("json=%s", filename), fmt.Sprintf("registry:%s", image)) // nolint: gosec
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = append(cmd.Env, fmt.Sprintf("SYFT_REGISTRY_AUTH_TOKEN=%s", authorization))
	if strings.HasPrefix(p.Config.HTTP.InternalEndpoint, "http://") {
		cmd.Env = append(cmd.Env, "SYFT_REGISTRY_INSECURE_USE_HTTP=true")
	}
	if strings.HasPrefix(p.Config.HTTP.InternalEndpoint, "https://") {
		cmd.Env = append(cmd.Env, "SYFT_REGISTRY_INSECURE_SKIP_TLS_VERIFY=true")
	}

	slog.Info("start sbom artifact", "artifactDigest", artifact.Digest, "env", cmd.Env, "cmd", cmd.String())

	defer func() {
		if utils.IsFile(filename) {
			e := os.Remove(filename)
			if e != nil {
				slog.Warn("remove file failed", "err", e)
			}
		}
	}()

	err = cmd.Run()
	if err != nil {
		var stdoutBytes = stdout.Bytes()
		var stderrBytes = stderr.Bytes()
		slog.Error("run syft failed", "err", err, "stdout", string(stdoutBytes), "stderr", string(stderrBytes))
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
		slog.Error("open sbom file failed", "err", err, "filename", filename)
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
		slog.Error("decode sbom file failed", "err", err, "filename", filename)
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
		slog.Error("marshal report failed", "err", err)
		statusChan <- decoratorArtifactStatus{Daemon: enums.DaemonVulnerability, Status: enums.TaskCommonStatusFailed, Message: err.Error()}
		return err
	}

	compressed, err := compress.Compress(filename)
	if err != nil {
		slog.Error("compress file failed", "err", err)
		statusChan <- decoratorArtifactStatus{Daemon: enums.DaemonSbom, Status: enums.TaskCommonStatusFailed, Message: err.Error()}
		return err
	}

	slog.Info("success sbom artifact", "artifactDigest", artifact.Digest)

	statusChan <- decoratorArtifactStatus{Daemon: enums.DaemonSbom,
		Status:  enums.TaskCommonStatusSuccess,
		Message: "", Raw: compressed, Result: reportBytes}

	return nil
}
