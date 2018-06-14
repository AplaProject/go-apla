package api

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/service"
	"github.com/GenesisKernel/go-genesis/packages/statsd"
	"github.com/gorilla/mux"

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

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger := getLogger(r)
				logger.WithFields(log.Fields{
					"type":  consts.PanicRecoveredError,
					"error": err,
					"stack": string(debug.Stack()),
				}).Error("panic recovered error")

				fmt.Println("API Recovered", fmt.Sprintf("%s: %s", r, debug.Stack()))
				errorResponse(w, errRecovered)
			}
		}()

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

		errorResponse(w, reason)
	})
}

func StatsdMiddleware(next http.Handler) http.Handler {
	const v = 1.0

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)

		counterName := statsd.APIRouteCounterName(r.Method, route.GetName())
		statsd.Client.Inc(counterName+statsd.Count, 1, v)
		startTime := time.Now()

		defer func() {
			statsd.Client.TimingDuration(counterName+statsd.Time, time.Since(startTime), v)
		}()

		next.ServeHTTP(w, r)
	})
}
