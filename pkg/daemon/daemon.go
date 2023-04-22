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
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/logger"
)

//go:generate go-enum

// Daemon x ENUM(
// Vulnerability,
// Sbom,
// )
type Daemon string

// tasks all daemon tasks
var tasks = map[Daemon]func(context.Context, *asynq.Task) error{}

// topics all daemon topics
var topics = map[Daemon]string{
	DaemonSbom:          consts.TopicSbom,
	DaemonVulnerability: consts.TopicVulnerability,
}

// asynqCli asynq client
var asynqCli *asynq.Client

// RegisterTask registers a daemon task
func RegisterTask(name Daemon, handler func(context.Context, *asynq.Task) error) error {
	_, ok := tasks[name]
	if ok {
		return fmt.Errorf("daemon task %q already registered", name)
	}
	tasks[name] = handler
	return nil
}

// Initialize initializes the daemon tasks
func Initialize() error {
	redisOpt, err := asynq.ParseRedisURI(viper.GetString("redis.url"))
	if err != nil {
		return fmt.Errorf("asynq.ParseRedisURI error: %v", err)
	}
	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			Logger: &logger.Logger{},
		},
	)

	mux := asynq.NewServeMux()
	for taskType, handler := range tasks {
		topic, ok := topics[taskType]
		if !ok {
			return fmt.Errorf("topic for daemon task %q not found", taskType)
		}
		mux.HandleFunc(topic, handler)
	}

	go func() {
		err := srv.Run(mux)
		if err != nil {
			log.Fatal().Err(err).Msg("srv.Run error")
		}
	}()

	return nil
}

// InitializeClient initializes the daemon client
func InitializeClient() error {
	redisOpt, err := asynq.ParseRedisURI(viper.GetString("redis.url"))
	if err != nil {
		return fmt.Errorf("asynq.ParseRedisURI error: %v", err)
	}
	asynqCli = asynq.NewClient(redisOpt)
	return nil
}

// Enqueue enqueues a task
func Enqueue(topic string, payload []byte) error {
	task := asynq.NewTask(topic, payload)
	_, err := asynqCli.Enqueue(task)
	if err != nil {
		return fmt.Errorf("asynqCli.Enqueue error: %v", err)
	}
	return nil
}
