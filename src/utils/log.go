package utils

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"strings"

	"pet_adopter/src/config"
)

func GetFunctionName() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	values := strings.Split(frame.Function, "/")

	return values[len(values)-1]
}

func GetLoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(config.LoggerContextKey).(*slog.Logger); ok {
		return logger
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	logger.Error("failed to logger from context, new logger was created")

	return logger
}
