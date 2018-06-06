package api

import (
	"encoding/json"
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/service"

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

func NodeStateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reason errorType

		switch service.NodePauseType() {
		case service.NoPause:
			next.ServeHTTP(w, r)
			return
		case service.PauseTypeUpdatingBlockchain:
			reason = errUpdating
			break
		case service.PauseTypeStopingNetwork:
			reason = errStopping
			break
		}

		errorResponse(w, reason, http.StatusServiceUnavailable)
	})
}

func jsonResponse(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
}
