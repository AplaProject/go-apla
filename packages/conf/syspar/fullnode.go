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

package syspar

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	log "github.com/sirupsen/logrus"
)

const publicKeyLength = 64

var (
	errFullNodeInvalidValues = errors.New("Invalid values of the full_node parameter")
)

//because of PublicKey is byte
type fullNodeJSON struct {
	TCPAddress string `json:"tcp_address"`
	APIAddress string `json:"api_address"`
	//	KeyID      json.Number `json:"key_id"`
	PublicKey string      `json:"public_key"`
	UnbanTime json.Number `json:"unban_time,er"`
	Stopped   bool        `json:"stopped"`
}

// FullNode is storing full node data
type FullNode struct {
	TCPAddress string
	APIAddress string
	PublicKey  []byte
	UnbanTime  time.Time
	Stopped    bool
}

// UnmarshalJSON is custom json unmarshaller
func (fn *FullNode) UnmarshalJSON(b []byte) (err error) {
	data := fullNodeJSON{}
	if err = json.Unmarshal(b, &data); err != nil {
		log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err, "value": string(b)}).Error("Unmarshalling full nodes to json")
		return err
	}

	fn.TCPAddress = data.TCPAddress
	fn.APIAddress = data.APIAddress
	fn.Stopped = data.Stopped
	if fn.PublicKey, err = crypto.HexToPub(data.PublicKey); err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err, "value": data.PublicKey}).Error("converting full nodes public key from hex")
		return err
	}
	fn.UnbanTime = time.Unix(converter.StrToInt64(data.UnbanTime.String()), 0)

	if err = fn.Validate(); err != nil {
		return err
	}

	return nil
}

func (fn *FullNode) MarshalJSON() ([]byte, error) {
	jfn := fullNodeJSON{
		TCPAddress: fn.TCPAddress,
		APIAddress: fn.APIAddress,
		PublicKey:  crypto.PubToHex(fn.PublicKey),
		UnbanTime:  json.Number(strconv.FormatInt(fn.UnbanTime.Unix(), 10)),
	}

	data, err := json.Marshal(jfn)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("Marshalling full nodes to json")
		return nil, err
	}

	return data, nil
}

// ValidateURL returns error if the URL is invalid
func validateURL(rawurl string) error {
	u, err := url.ParseRequestURI(rawurl)
	if err != nil {
		return err
	}

	if len(u.Scheme) == 0 {
		return fmt.Errorf("Invalid scheme: %s", rawurl)
	}

	if len(u.Host) == 0 {
		return fmt.Errorf("Invalid host: %s", rawurl)
	}

	return nil
}

// Validate checks values
func (fn *FullNode) Validate() error {
	if len(fn.PublicKey) != publicKeyLength || len(fn.TCPAddress) == 0 {
		return errFullNodeInvalidValues
	}

	if err := validateURL(fn.APIAddress); err != nil {
		return err
	}

	return nil
}
