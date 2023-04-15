package daemon

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/ximager/ximager/pkg/logger"
)

// tasks all daemon tasks
var tasks = map[string]func(context.Context, *asynq.Task) error{}

// asynqCli asynq client
var asynqCli *asynq.Client

// RegisterTask registers a daemon task
func RegisterTask(name string, handler func(context.Context, *asynq.Task) error) error {
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
		mux.HandleFunc(taskType, handler)
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
