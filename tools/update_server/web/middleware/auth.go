package middleware

import (
	"net/http"

	"github.com/go-chi/render"
)

// Auth is basic auth middleware
func Auth(login string, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

			un, pw, authOK := r.BasicAuth()
			if authOK == false || un != login || pw != password {
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, struct{}{})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
