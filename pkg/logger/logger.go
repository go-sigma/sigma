package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// SetLevel sets the log level
func SetLevel(level int) {
	if level < int(zerolog.TraceLevel) || level > int(zerolog.FatalLevel) {
		level = int(zerolog.InfoLevel)
	}

	var timeFormat = "2006-01-02 15:04:05" // change it to 'time.DataTime' om go 1.20
	zerolog.SetGlobalLevel(zerolog.Level(level))
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: timeFormat, FormatCaller: func(i interface{}) string {
		var c string
		if cc, ok := i.(string); ok {
			c = cc
		}
		if len(c) > 0 && strings.Contains(c, "/") {
			lastIndex := strings.LastIndex(c, "/")
			c = c[lastIndex+1:]
		}
		return c
	}}).With().Caller().Timestamp().Logger()
}

// Logger is a wrapper for zerolog
type Logger struct{}

// Debug logs a message at Debug level.
func (l *Logger) Debug(args ...interface{}) {
	log.Debug().Msg(fmt.Sprintf("%v", args...))
}

// Info logs a message at Info level.
func (l *Logger) Info(args ...interface{}) {
	log.Info().Msg(fmt.Sprintf("%v", args...))
}

// Warn logs a message at Warning level.
func (l *Logger) Warn(args ...interface{}) {
	log.Warn().Msg(fmt.Sprintf("%v", args...))
}

// Error logs a message at Error level.
func (l *Logger) Error(args ...interface{}) {
	log.Error().Msg(fmt.Sprintf("%v", args...))
}

// Fatal logs a message at Fatal level
// and process will exit with status set to 1.
func (l *Logger) Fatal(args ...interface{}) {
	log.Fatal().Msg(fmt.Sprintf("%v", args...))
}
