package logging

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitLogger() {
	// Configure the logger
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	output.FormatLevel = func(i interface{}) string {
		return fmt.Sprintf("| \x1b[%dm%s\x1b[0m |", levelColor(i), strings.ToUpper(i.(string)))
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("=== %s ===", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("\x1b[%dm%s\x1b[0m:", fieldColor(i), i)
	}
	output.FormatFieldValue = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}

	log.Logger = log.Output(output).With().Timestamp().Logger()
}

func levelColor(level interface{}) int {
	switch level {
	case "debug":
		return 36
	case "info":
		return 32
	case "warn":
		return 33
	case "error":
		return 31
	case "fatal":
		return 35
	case "panic":
		return 35
	default:
		return 0
	}
}

func fieldColor(field interface{}) int {
	switch field {
	case "dropletID", "name", "ip", "subdomain", "domain", "keyID", "count":
		return 34
	default:
		return 0
	}
}