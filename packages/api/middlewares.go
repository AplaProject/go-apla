package api

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.WithFields(log.Fields{
			"headers":  r.Header,
			"path":     r.URL.Path,
			"protocol": r.Proto,
			"remote":   r.RemoteAddr,
		})
		logger.Info("received http request")

		r = setLogger(r, logger)

		next.ServeHTTP(w, r)
	})
}

func jsonResponse(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
}
