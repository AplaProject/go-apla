package api

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/service"
	"github.com/GenesisKernel/go-genesis/packages/statsd"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func authRequire(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		client := getClient(r)
		if client != nil && client.KeyID != 0 {
			next(w, r)
			return
		}

		logger := getLogger(r)
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("wallet is empty")
		errorResponse(w, errUnauthorized)
	}
}

func loggerMiddleware(next http.Handler) http.Handler {
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

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger := getLogger(r)
				logger.WithFields(log.Fields{
					"type":  consts.PanicRecoveredError,
					"error": err,
					"stack": string(debug.Stack()),
				}).Error("panic recovered error")
				fmt.Println("API Recovered", fmt.Sprintf("%s: %s", err, debug.Stack()))
				errorResponse(w, errRecovered)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func nodeStateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reason errType
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

func tokenMiddleware(next http.Handler) http.Handler {
	const authHeader = "AUTHORIZATION"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := parseJWTToken(r.Header.Get(authHeader))
		if err != nil {
			logger := getLogger(r)
			logger.WithFields(log.Fields{"type": consts.JWTError, "error": err}).Error("starting session")
			if err, ok := err.(jwt.ValidationError); ok {
				if (err.Errors & jwt.ValidationErrorExpired) != 0 {
					errorResponse(w, errTokenExpired.Errorf(err.Error()))
					return
				}
			}
			errorResponse(w, err, http.StatusBadRequest)
			return
		}
		if token != nil && token.Valid {
			r = setToken(r, token)
		}
		next.ServeHTTP(w, r)
	})
}

func clientMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := getToken(r)
		var client *Client
		if token != nil { // get client from token
			var err error
			if client, err = getClientFromToken(token); err != nil {
				errorResponse(w, err)
				return
			}
		}
		if client == nil {
			// create client with default ecosystem
			client = &Client{EcosystemID: 1}
		}
		r = setClient(r, client)
		next.ServeHTTP(w, r)
	})
}

func statsdMiddleware(next http.Handler) http.Handler {
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
