package log

import (
	"io"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/nekrassov01/logger/integrations/awssdk"
	"github.com/nekrassov01/logger/log"
)

// Logger is a global logger instance used across the application.
var Logger = log.NewLogger(log.NewCLIHandler(io.Discard))

// SetAppLogger configures the global Logger with the specified log level and output writer.
func SetAppLogger(w io.Writer, level string) {
	Logger = log.NewLogger(setHandler(w, level, "IGNORESYNC"))
}

// SetSDKLogger configures the AWS SDK logger with the specified log level and output writer.
func SetSDKLogger(w io.Writer, level string, cfg *aws.Config) {
	cfg.Logger = awssdk.NewLogger(setHandler(w, level, "SDK"))
	cfg.ClientLogMode = aws.LogRequest |
		aws.LogResponse |
		aws.LogRetries |
		aws.LogSigning |
		aws.LogDeprecatedUsage
}

// setHandler creates a slog.Handler with specified configuration.
func setHandler(w io.Writer, level, label string) slog.Handler {
	var lv slog.Level
	if err := lv.UnmarshalText([]byte(level)); err != nil {
		lv = slog.LevelInfo
	}
	style := log.Style2()
	style.Caller.Fullpath = true
	return log.NewCLIHandler(
		w,
		log.WithLabel(label),
		log.WithLevel(lv),
		log.WithCaller(lv <= slog.LevelDebug),
		log.WithStyle(style),
		log.WithTime(true),
	)
}
