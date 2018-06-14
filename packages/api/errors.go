// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package api

import (
	"fmt"
	"net/http"
)

var (
	errContract        = errorType{"E_CONTRACT", "There is not %s contract", http.StatusBadRequest}
	errDBNil           = errorType{"E_DBNIL", "DB is nil", http.StatusBadRequest}
	errDeletedKey      = errorType{"E_DELETEDKEY", "The key is deleted", http.StatusForbidden}
	errEcosystem       = errorType{"E_ECOSYSTEM", "Ecosystem %d doesn't exist", http.StatusBadRequest}
	errEmptyPublic     = errorType{"E_EMPTYPUBLIC", "Public key is undefined", http.StatusBadRequest}
	errEmptySign       = errorType{"E_EMPTYSIGN", "Signature is undefined", http.StatusBadRequest}
	errHashWrong       = errorType{"E_HASHWRONG", "Hash is incorrect", http.StatusBadRequest}
	errHashNotFound    = errorType{"E_HASHNOTFOUND", "Hash has not been found", http.StatusBadRequest}
	errHeavyPage       = errorType{"E_HEAVYPAGE", "This page is heavy", http.StatusInternalServerError}
	errInstalled       = errorType{"E_INSTALLED", "Apla is already installed", http.StatusBadRequest}
	errInvalidWallet   = errorType{"E_INVALIDWALLET", "Wallet %s is not valid", http.StatusBadRequest}
	errNotFound        = errorType{"E_NOTFOUND", "Page not found", http.StatusNotFound}
	errNotInstalled    = errorType{"E_NOTINSTALLED", "Apla is not installed", http.StatusBadRequest}
	errParamNotFound   = errorType{"E_PARAMNOTFOUND", "Parameter %s has not been found", http.StatusNotFound}
	errPermission      = errorType{"E_PERMISSION", "Permission denied", http.StatusBadRequest}
	errQuery           = errorType{"E_QUERY", "DB query is wrong", http.StatusInternalServerError}
	errRecovered       = errorType{"E_RECOVERED", "API recovered", http.StatusInternalServerError}
	errRefreshToken    = errorType{"E_REFRESHTOKEN", "Refresh token is not valid", http.StatusBadRequest}
	errServer          = errorType{"E_SERVER", "Server error", http.StatusInternalServerError}
	errSignature       = errorType{"E_SIGNATURE", "Signature is incorrect", http.StatusBadRequest}
	errUnknownSign     = errorType{"E_UNKNOWNSIGN", "Unknown signature", http.StatusBadRequest}
	errStateLogin      = errorType{"E_STATELOGIN", "%s is not a membership of ecosystem %s", http.StatusBadRequest}
	errTableNotFound   = errorType{"E_TABLENOTFOUND", "Table %s has not been found", http.StatusBadRequest}
	errToken           = errorType{"E_TOKEN", "Token is not valid", http.StatusBadRequest}
	errTokenExpired    = errorType{"E_TOKENEXPIRED", "Token is expired by %s", http.StatusUnauthorized}
	errUnauthorized    = errorType{"E_UNAUTHORIZED", "Unauthorized", http.StatusUnauthorized}
	errUndefineVal     = errorType{"E_UNDEFINEVAL", "Value %s is undefined", http.StatusBadRequest}
	errUnknownUID      = errorType{"E_UNKNOWNUID", "Unknown uid", http.StatusBadRequest}
	errVDE             = errorType{"E_VDE", "Virtual Dedicated Ecosystem %d doesn't exist", http.StatusBadRequest}
	errVDECreated      = errorType{"E_VDECREATED", "Virtual Dedicated Ecosystem is already created", http.StatusBadRequest}
	errRequestNotFound = errorType{"E_REQUESTNOTFOUND", "Request %s doesn't exist", http.StatusNotFound}
	errCheckRole       = errorType{"E_CHECKROLE", "Check role", http.StatusBadRequest}
	errUpdating        = errorType{"E_UPDATING", "Node is updating blockchain", http.StatusServiceUnavailable}
	errStopping        = errorType{"E_STOPPING", "Network is stopping", http.StatusServiceUnavailable}
)

type errorType struct {
	Err     string `json:"error"`
	Message string `json:"msg"`
	Status  int    `json:"-"`
}

func (et errorType) Error() string {
	return string(et.Err)
}

// Errof returns formating error
func (et errorType) Errorf(v ...interface{}) errorType {
	et.Message = fmt.Sprintf(et.Message, v...)
	return et
}

func newError(err error, status int, v ...interface{}) errorType {
	et, ok := err.(errorType)
	if ok {
		et = et.Errorf(v...)
	} else {
		et = errServer
		et.Message = err.Error()
	}
	et.Status = status
	return et
}

func errorResponse(w http.ResponseWriter, err error) {
	et, ok := err.(errorType)
	if !ok {
		et = errServer
		et.Message = err.Error()
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(et.Status)

	jsonResponse(w, et)
}
