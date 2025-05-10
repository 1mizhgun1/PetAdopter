package middleware

import (
	"context"
	goerrors "errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"pet_adopter/src/config"
	"pet_adopter/src/user/logic"
	"pet_adopter/src/utils"
)

const msgNoAuth = "no auth"

func hasAuth(r *http.Request, sessionLogic *logic.SessionLogic, cfg config.SessionConfig) (int, func(), bool) {
	ctx := r.Context()
	username := r.URL.Query().Get("username")

	headerToken := r.Header.Get("Authorization")
	if !strings.HasPrefix(headerToken, "Bearer ") {
		return http.StatusUnauthorized, func() { utils.LogErrorMessage(ctx, "invalid token in Authorization header") }, false
	}
	headerToken = strings.TrimPrefix(headerToken, "Bearer ")

	cookieToken, err := r.Cookie(cfg.AccessTokenCookieName)
	if err != nil {
		if goerrors.Is(err, http.ErrNoCookie) {
			return http.StatusUnauthorized, func() { utils.LogErrorMessage(ctx, "no session cookie") }, false
		}
		return http.StatusInternalServerError, func() { utils.LogError(ctx, err, "failed to get access token from cookie") }, false
	}

	if cookieToken.Value != headerToken {
		return http.StatusUnauthorized, func() { utils.LogErrorMessage(ctx, "tokens are different") }, false
	}

	correctSession, err := sessionLogic.CheckSession(ctx, username, headerToken)
	if err != nil {
		return http.StatusInternalServerError, func() { utils.LogError(ctx, err, "failed to check session") }, false
	}
	if !correctSession {
		return http.StatusUnauthorized, func() { utils.LogErrorMessage(ctx, "invalid session") }, false
	}

	return http.StatusOK, func() {}, true
}

func CreateSessionMiddleware(userLogic *logic.UserLogic, sessionLogic *logic.SessionLogic, cfg config.SessionConfig, needAuth bool) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v1/ads" && r.URL.Query().Get("radius") != "" {
				needAuth = true
			}

			status, logFunc, auth := hasAuth(r, sessionLogic, cfg)
			if status == http.StatusInternalServerError {
				logFunc()
				http.Error(w, utils.Internal, status)
				return
			}
			if !auth && needAuth {
				logFunc()
				http.Error(w, msgNoAuth, status)
				return
			}

			if auth {
				username := r.URL.Query().Get("username")
				userData, err := userLogic.GetUserByUsername(r.Context(), username)
				if err != nil {
					utils.LogError(r.Context(), err, "failed to get user by username")
					http.Error(w, utils.Internal, http.StatusInternalServerError)
					return
				}

				r = r.WithContext(context.WithValue(r.Context(), config.UserIDContextKey, userData.ID))
				r = r.WithContext(context.WithValue(r.Context(), config.UsernameContextKey, userData.Username))

				utils.LogInfoMessage(r.Context(), fmt.Sprintf("user %s authenticated", username))
			} else {
				utils.LogInfoMessage(r.Context(), "user is not authenticated")
			}

			next.ServeHTTP(w, r)
		})
	}
}
