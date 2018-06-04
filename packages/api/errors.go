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

type errorType string

func (et errorType) Error() string {
	return string(et)
}

const (
	errContract        errorType = "E_CONTRACT"
	errDBNil           errorType = "E_DBNIL"
	errDeletedKey      errorType = "E_DELETEDKEY"
	errEcosystem       errorType = "E_ECOSYSTEM"
	errEmptyPublic     errorType = "E_EMPTYPUBLIC"
	errEmptySign       errorType = "E_EMPTYSIGN"
	errHashWrong       errorType = "E_HASHWRONG"
	errHashNotFound    errorType = "E_HASHNOTFOUND"
	errHeavyPage       errorType = "E_HEAVYPAGE"
	errInstalled       errorType = "E_INSTALLED"
	errInvalidWallet   errorType = "E_INVALIDWALLET"
	errNotFound        errorType = "E_NOTFOUND"
	errNotInstalled    errorType = "E_NOTINSTALLED"
	errParamNotFound   errorType = "E_PARAMNOTFOUND"
	errPermission      errorType = "E_PERMISSION"
	errQuery           errorType = "E_QUERY"
	errRecovered       errorType = "E_RECOVERED"
	errRefreshToken    errorType = "E_REFRESHTOKEN"
	errServer          errorType = "E_SERVER"
	errSignature       errorType = "E_SIGNATURE"
	errUnknownSign     errorType = "E_UNKNOWNSIGN"
	errStateLogin      errorType = "E_STATELOGIN"
	errTableNotFound   errorType = "E_TABLENOTFOUND"
	errToken           errorType = "E_TOKEN"
	errTokenExpired    errorType = "E_TOKENEXPIRED"
	errUnauthorized    errorType = "E_UNAUTHORIZED"
	errUndefineVal     errorType = "E_UNDEFINEVAL"
	errUnknownUID      errorType = "E_UNKNOWNUID"
	errVDE             errorType = "E_VDE"
	errVDECreated      errorType = "E_VDECREATED"
	errRequestNotFound errorType = "E_REQUESTNOTFOUND"
	errCheckRole       errorType = "E_CHECKROLE"
	errUpdating        errorType = "E_UPDATING"
	errStopping        errorType = "E_STOPPING"
)

var (
	errorDescriptions = map[errorType]string{
		errContract:        "There is not %s contract",
		errDBNil:           "DB is nil",
		errDeletedKey:      "The key is deleted",
		errEcosystem:       "Ecosystem %d doesn't exist",
		errEmptyPublic:     "Public key is undefined",
		errEmptySign:       "Signature is undefined",
		errHashWrong:       "Hash is incorrect",
		errHashNotFound:    "Hash has not been found",
		errHeavyPage:       "This page is heavy",
		errInstalled:       "Apla is already installed",
		errInvalidWallet:   "Wallet %s is not valid",
		errNotFound:        "Page not found",
		errNotInstalled:    "Apla is not installed",
		errParamNotFound:   "Parameter %s has not been found",
		errPermission:      "Permission denied",
		errQuery:           "DB query is wrong",
		errRecovered:       "API recovered",
		errRefreshToken:    "Refresh token is not valid",
		errServer:          "Server error",
		errSignature:       "Signature is incorrect",
		errUnknownSign:     "Unknown signature",
		errStateLogin:      "%s is not a membership of ecosystem %s",
		errTableNotFound:   "Table %s has not been found",
		errToken:           "Token is not valid",
		errTokenExpired:    "Token is expired by %s",
		errUnauthorized:    "Unauthorized",
		errUndefineVal:     "Value %s is undefined",
		errUnknownUID:      "Unknown uid",
		errVDE:             "Virtual Dedicated Ecosystem %d doesn't exist",
		errVDECreated:      "Virtual Dedicated Ecosystem is already created",
		errRequestNotFound: "Request %s doesn't exist",
		errCheckRole:       "Check role",
		errUpdating:        "Node is updating blockchain",
		errStopping:        "Network is stopping",
	}

	//TODO: remove
	apiErrors = map[string]string{}
)

type errResult struct {
	Error   errorType `json:"error"`
	Message string    `json:"msg"`
}

func errorResponse(w http.ResponseWriter, err interface{}, code int, params ...interface{}) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)

	result := &errResult{}

	switch v := err.(type) {
	case errorType:
		result.Error = v
		result.Message = fmt.Sprintf(errorDescriptions[v], params...)
	case interface{}:
		result.Error = errServer
		if err, ok := v.(error); ok {
			result.Message = err.Error()
		}
	}

	jsonResponse(w, result)
}
