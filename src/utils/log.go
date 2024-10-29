package utils

import (
	"context"
	"log/slog"
	"os"

	"github.com/pkg/errors"
	"pet_adopter/src/config"
)

const (
	Internal = "internal"
	Invalid  = "invalid"

	MsgErrMarshalResponse  = "failed to unmarshal request"
	MsgErrUnmarshalRequest = "failed to unmarshal request"
)

func GetLoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(config.LoggerContextKey).(*slog.Logger); ok {
		return logger
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	logger.Error("failed to logger from context, new logger was created")

	return logger
}

func LogError(ctx context.Context, err error, msg string) {
	logger := GetLoggerFromContext(ctx)
	logger.Error(errors.Wrap(err, msg).Error())
}

func LogNotAnAdminError(ctx context.Context) {
	logger := GetLoggerFromContext(ctx)
	logger.Error("not an admin")
}
