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

package smart

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"

	log "github.com/sirupsen/logrus"
)

func logError(err error, errType string, comment string) error {
	log.WithFields(log.Fields{"type": errType, "error": err}).Error(comment)
	return err
}

func logErrorf(pattern string, param interface{}, errType string, comment string) error {
	err := fmt.Errorf(pattern, param)
	log.WithFields(log.Fields{"type": errType, "error": err}).Error(comment)
	return err
}

func logErrorShort(err error, errType string) error {
	return logError(err, errType, err.Error())
}

func logErrorfShort(pattern string, param interface{}, errType string) error {
	return logErrorShort(fmt.Errorf(pattern, param), errType)
}

func logErrorValue(err error, errType string, comment, value string) error {
	log.WithFields(log.Fields{"type": errType, "error": err, "value": value}).Error(comment)
	return err
}

func logErrorDB(err error, comment string) error {
	return logError(err, consts.DBError, comment)
}

func unmarshalJSON(input []byte, v interface{}, comment string) (err error) {
	if err = json.Unmarshal(input, v); err != nil {
		return logErrorValue(err, consts.JSONUnmarshallError, comment, string(input))
	}
	return nil
}

func marshalJSON(v interface{}, comment string) (out []byte, err error) {
	out, err = json.Marshal(v)
	if err != nil {
		logError(err, consts.JSONMarshallError, comment)
	}
	return
}

func validateAccess(funcName string, sc *SmartContract, contracts ...string) error {
	if !accessContracts(sc, contracts...) {
		err := fmt.Errorf(eAccessContract, funcName, strings.Join(contracts, ` or `))
		return logError(err, consts.IncorrectCallingContract, err.Error())
	}
	return nil
}
