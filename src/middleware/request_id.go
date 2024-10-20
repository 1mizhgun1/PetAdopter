package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/satori/uuid"
	"pet_adopter/src/config"
)

func CreateRequestIDMiddleware(logger *slog.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := uuid.NewV4().String()

			ctx := context.WithValue(
				r.Context(),
				config.LoggerContextKey,
				logger.With(slog.String("x-request-id", reqID)),
			)

			r = r.WithContext(ctx)
			w.Header().Set("X-Request-ID", reqID)

			logger.Info(fmt.Sprintf("input request %s uri=%s", reqID, r.RequestURI))
			next.ServeHTTP(w, r)
		})
	}
}
