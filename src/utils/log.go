package utils

import (
	"context"
	"log/slog"
	"os"

	"github.com/pkg/errors"
	"github.com/satori/uuid"
	"pet_adopter/src/config"
)

const (
	Internal = "internal"
	Invalid  = "invalid"
	NotFound = "not_found"

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

func GetUserIDFromContext(ctx context.Context) uuid.UUID {
	if userID, ok := ctx.Value(config.UserIDContextKey).(uuid.UUID); ok {
		return userID
	}

	return uuid.Nil
}

func GetUsernameFromContext(ctx context.Context) string {
	if username, ok := ctx.Value(config.UsernameContextKey).(string); ok {
		return username
	}

	return ""
}

func LogError(ctx context.Context, err error, msg string) {
	logger := GetLoggerFromContext(ctx)
	logger.Error(errors.Wrap(err, msg).Error())
}

func LogNotAnAdminError(ctx context.Context) {
	logger := GetLoggerFromContext(ctx)
	logger.Error("not an admin")
}

func LogErrorMessage(ctx context.Context, msg string) {
	logger := GetLoggerFromContext(ctx)
	logger.Error(msg)
}

func LogInfoMessage(ctx context.Context, msg string) {
	logger := GetLoggerFromContext(ctx)
	logger.Info(msg)
}
