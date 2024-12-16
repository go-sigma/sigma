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

package cmd

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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/inits"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/modules/locker"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/token"
)

func init() {
	toolsCmd.AddCommand(
		toolsMiddlewareCheckerCmd(),
		toolsForPushBuilderImageCmd(),
	)

	rootCmd.AddCommand(toolsCmd)
}

// toolsCmd represents the tools command
var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Tools for sigma",
}

func toolsForPushBuilderImageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push-builder-images",
		Short: "Push builder images to distribution",
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			initConfig()
			logger.SetLevel(viper.GetString("log.level"))
		},
		Run: func(_ *cobra.Command, _ []string) {
			// err := configs.Initialize()
			// if err != nil {
			// 	log.Error().Err(err).Msg("initialize configs with error")
			// 	return
			// }

			digCon, err := inits.NewDigContainer()
			if err != nil {
				log.Error().Err(err).Msg("new dig container failed")
				return
			}

			err = dal.Initialize(digCon)
			if err != nil {
				log.Error().Err(err).Msg("initialize database with error")
				return
			}

			err = initBaseimage(digCon)
			if err != nil {
				log.Error().Err(err).Msg("push builder image with error")
				return
			}
		},
	}

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/sigma/sigma.yaml)")

	return cmd
}

func initBaseimage(digCon *dig.Container) error {
	config := utils.MustGetObjFromDigCon[configs.Configuration](digCon)
	if !config.Daemon.Builder.Enabled {
		return nil
	}
	dir := strings.TrimPrefix(consts.BuilderImagePath, "./")
	if !utils.IsDir(dir) {
		log.Info().Msg("builder image not found, skip push image")
		return nil
	}
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()
	err := locker.Locker.AcquireWithRenew(ctx, consts.LockerBaseimage, time.Second*3, time.Second*5)
	if err != nil {
		return err
	}

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
				err := pushImage(digCon, path, name, version)
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

func pushImage(digCon *dig.Container, path, name, version string) error {
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
	tokenService, err := token.New(digCon)
	if err != nil {
		return err
	}
	config := utils.MustGetObjFromDigCon[configs.Configuration](digCon)
	autoToken, err := tokenService.New(userObj.ID, config.Auth.Jwt.Ttl)
	if err != nil {
		return err
	}
	cmd := exec.Command("skopeo", "--insecure-policy", "copy", "--dest-registry-token", autoToken, "--dest-tls-verify=false", "-a", fmt.Sprintf("oci-archive:%s", path), fmt.Sprintf("docker://%s/library/%s:latest", utils.TrimHTTP(config.HTTP.InternalEndpoint), name)) // nolint: gosec
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

func toolsMiddlewareCheckerCmd() *cobra.Command {
	var waitTimeout time.Duration
	cmd := &cobra.Command{
		Use:   "middleware-checker",
		Short: "Check all of middleware status all ready",
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			initConfig()
			logger.SetLevel(viper.GetString("log.level"))
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			// err := configs.Initialize()
			// if err != nil {
			// 	log.Error().Err(err).Msg("initialize configs with error")
			// 	return err
			// }

			if waitTimeout == 0 {
				waitTimeout = time.Second * 120
			}

			ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
			defer cancel()

			for {
				select {
				case <-ctx.Done():
					return fmt.Errorf("middleware checker timeout, not all of middleware ready")
				case <-time.After(time.Second * 3):
					err := configs.CheckMiddleware()
					if err != nil {
						log.Error().Err(err).Msg("check middleware with error")
					} else {
						return nil
					}
				}
			}
		},
	}

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/sigma/sigma.yaml)")
	cmd.PersistentFlags().DurationVar(&waitTimeout, "wait-timeout", time.Second*120, "wait middleware timeout")

	return cmd
}
