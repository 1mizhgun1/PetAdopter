package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"pet_adopter/src/utils"
)

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recoverLogger := utils.GetLoggerFromContext(r.Context()).With(slog.String("func", utils.GetFunctionName()))

		defer func() {
			if err := recover(); err != nil {
				recoverLogger.Error(fmt.Sprintf("panic recovered: %v", err))
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
