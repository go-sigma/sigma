package logger

import (
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
