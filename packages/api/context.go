// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package api

import (
	"context"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

type contextKey int

const (
	contextKeyLogger contextKey = iota
	contextKeyToken
	contextKeyClient
)

func setContext(r *http.Request, key, value interface{}) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), key, value))
}

func getContext(r *http.Request, key interface{}) interface{} {
	return r.Context().Value(key)
}

func setLogger(r *http.Request, log *log.Entry) *http.Request {
	return setContext(r, contextKeyLogger, log)
}

func getLogger(r *http.Request) *log.Entry {
	if v := getContext(r, contextKeyLogger); v != nil {
		return v.(*log.Entry)
	}
	return nil
}

func setToken(r *http.Request, token *jwt.Token) *http.Request {
	return setContext(r, contextKeyToken, token)
}

func getToken(r *http.Request) *jwt.Token {
	if v := getContext(r, contextKeyToken); v != nil {
		return v.(*jwt.Token)
	}
	return nil
}

func setClient(r *http.Request, client *Client) *http.Request {
	return setContext(r, contextKeyClient, client)
}

func getClient(r *http.Request) *Client {
	if v := getContext(r, contextKeyClient); v != nil {
		return v.(*Client)
	}
	return nil
}
