// Copyright 2024 sigma
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

package inits

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/modules/locker"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/token"
)

func init() {
	afterInit["baseimage"] = initBaseimage
}

const baseImageDir = "/baseimages"

func initBaseimage(config configs.Configuration) error {
	if !config.Daemon.Builder.Enabled {
		return nil
	}
	dir := strings.TrimPrefix(baseImageDir, "./")
	if !utils.IsDir(dir) {
		log.Info().Msg("Baseimage not found")
		return nil
	}
	locker, err := locker.New(config)
	if err != nil {
		return err
	}
	lock, err := locker.Lock(context.Background(), consts.LockerBaseimage, time.Second*30)
	if err != nil {
		return err
	}
	defer func() {
		err := lock.Unlock()
		if err != nil {
			log.Error().Err(err).Msg("Initialize baseimage failed")
		}
	}()

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.HasSuffix(path, ".tar") {
			d := strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(path, dir), "/"), ".tar")
			var version string
			var name string
			if strings.HasPrefix(d, "dockerfile.") {
				name = "dockerfile"
				version = strings.TrimPrefix(d, "dockerfile.")
			} else if strings.HasPrefix(d, "builder.") {
				name = "builder"
				version = strings.TrimPrefix(d, "builder.")
			}
			if version != "" {
				err := pushImage(config, path, name, version)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func pushImage(config configs.Configuration, path, name, version string) error {
	ctx := log.Logger.WithContext(context.Background())

	userService := dao.NewUserServiceFactory().New()
	settingService := dao.NewSettingServiceFactory().New()

	var key string
	switch name {
	case "dockerfile":
		key = consts.SettingBaseimageDockerfileKey
	case "builder":
		key = consts.SettingBaseimageBuilderKey
	default:
		return fmt.Errorf("name(%s) is not support", name)
	}

	versions, err := settingService.Get(ctx, key)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("get baseimage.dockerfile setting failed: %v", err)
	}
	var versionsVal string
	if versions != nil {
		versionsVal = string(versions.Val)
	}
	var sets = mapset.NewSet(strings.Split(versionsVal, ",")...)
	fmt.Println(versionsVal == "", versionsVal != "" && sets.ContainsOne(version), !(versionsVal == "" || (versionsVal != "" && sets.ContainsOne(version))))
	if !(versionsVal == "" || (versionsVal != "" && sets.ContainsOne(version))) {
		return nil
	}
	if !sets.Add(version) {
		return fmt.Errorf("add version to sets failed")
	}

	userObj, err := userService.GetByUsername(ctx, consts.UserInternal)
	if err != nil {
		return err
	}
	tokenService, err := token.NewTokenService(config.Auth.Jwt.PrivateKey)
	if err != nil {
		return err
	}
	authorization, err := tokenService.New(userObj.ID, config.Auth.Jwt.Ttl)
	if err != nil {
		return err
	}
	cmd := exec.Command("skopeo", "--insecure-policy", "copy", "--dest-registry-token", authorization, "--dest-tls-verify=false", "-a", fmt.Sprintf("oci-archive:%s", path), fmt.Sprintf("docker://%s/library/%s:%s", utils.TrimHTTP(config.HTTP.InternalEndpoint), name, version)) // nolint: gosec
	fmt.Println(cmd.String())
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	defer func() {
		if err != nil {
			if strings.TrimSpace(stdout.String()) != "" {
				log.Error().Err(err).Msgf("skopeo copy image failed stdout: %s", strings.TrimSpace(stdout.String()))
			}
			if strings.TrimSpace(stderr.String()) != "" {
				log.Error().Err(err).Msgf("skopeo copy image failed stderr: %s", strings.TrimSpace(stderr.String()))
			}
		} else {
			var val = strings.Join(sets.ToSlice(), ",")
			if versionsVal == "" {
				val = version
			}
			err := settingService.Update(ctx, key, []byte(val))
			if err != nil {
				log.Error().Err(err).Msg("update setting failed")
			}
		}
	}()
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
