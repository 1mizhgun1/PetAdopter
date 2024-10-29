package middleware

import (
	"net/http"
	"os"

	"pet_adopter/src/utils"
)

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("token") != os.Getenv("ADMIN_TOKEN") {
			utils.LogNotAnAdminError(r.Context())
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
