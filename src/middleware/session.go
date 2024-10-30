package middleware

import (
	"context"
	goerrors "errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"pet_adopter/src/config"
	"pet_adopter/src/user/logic"
	"pet_adopter/src/utils"
)

const (
	msgNoAuth = "no auth"
)

func CreateSessionMiddleware(userLogic *logic.UserLogic, sessionLogic *logic.SessionLogic, cfg config.SessionConfig) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username := r.URL.Query().Get("username")

			headerToken := r.Header.Get("Authorization")
			if !strings.HasPrefix(headerToken, "Bearer ") {
				utils.LogErrorMessage(r.Context(), "invalid token in Authorization header")
				http.Error(w, msgNoAuth, http.StatusBadRequest)
				return
			}
			headerToken = strings.TrimPrefix(headerToken, "Bearer ")

			cookieToken, err := r.Cookie(cfg.AccessTokenCookieName)
			if err != nil {
				if goerrors.Is(err, http.ErrNoCookie) {
					utils.LogErrorMessage(r.Context(), "no session cookie")
					http.Error(w, msgNoAuth, http.StatusBadRequest)
				} else {
					utils.LogError(r.Context(), err, "failed to get access token from cookie")
					http.Error(w, utils.Internal, http.StatusInternalServerError)
				}
				return
			}

			if cookieToken.Value != headerToken {
				utils.LogErrorMessage(r.Context(), "tokens are different")
				http.Error(w, msgNoAuth, http.StatusBadRequest)
				return
			}

			correctSession, err := sessionLogic.CheckSession(r.Context(), username, headerToken)
			if err != nil {
				utils.LogError(r.Context(), err, "failed to check session")
				http.Error(w, utils.Internal, http.StatusInternalServerError)
				return
			}
			if !correctSession {
				utils.LogErrorMessage(r.Context(), "invalid session")
				http.Error(w, msgNoAuth, http.StatusBadRequest)
				return
			}

			userData, err := userLogic.GetUserByUsername(r.Context(), username)
			if err != nil {
				utils.LogError(r.Context(), err, "failed to get user by username")
				http.Error(w, utils.Internal, http.StatusInternalServerError)
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), config.UserIDContextKey, userData.ID))
			r = r.WithContext(context.WithValue(r.Context(), config.UsernameContextKey, userData.Username))

			next.ServeHTTP(w, r)
		})
	}
}
