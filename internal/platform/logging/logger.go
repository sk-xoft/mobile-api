package logging

import (
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const defaultTimeZone = "Asia/Bangkok"

// New creates the process-wide structured logger. Production logs are JSON so
// they can be consumed consistently by local tools and deployment platforms.
func New(timeZone string) (*zap.Logger, error) {
	if strings.TrimSpace(timeZone) == "" {
		timeZone = defaultTimeZone
	}
	location, err := time.LoadLocation(timeZone)
	if err != nil {
		return nil, err
	}
	config := zap.NewProductionConfig()
	config.Encoding = "json"
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.EncoderConfig.EncodeTime = timeEncoder(location)
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	return config.Build(zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel), zap.ErrorOutput(os.Stderr))
}

func timeEncoder(location *time.Location) zapcore.TimeEncoder {
	return func(value time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(value.In(location).Format(time.RFC3339Nano))
	}
}

func Install(logger *zap.Logger) func() {
	undoGlobals := zap.ReplaceGlobals(logger)
	undoStdLog := zap.RedirectStdLog(logger)
	return func() {
		undoStdLog()
		undoGlobals()
	}
}
