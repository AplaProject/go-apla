package syspar

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/GenesisKernel/go-genesis/packages/converter"
)

const publicKeyLength = 64

var (
	errFullNodeInvalidFormat = errors.New("Invalid format of the full_node parameter")
	errFullNodeInvalidValues = errors.New("Invalid values of the full_node parameter")
)

// FullNode is storing full node data
type FullNode struct {
	TCPAddress string
	APIAddress string
	KeyID      int64
	PublicKey  []byte
}

func (fn *FullNode) UnmarshalJSON(b []byte) (err error) {
	var data []string
	if err = json.Unmarshal(b, &data); err != nil {
		return err
	}

	if len(data) != 4 {
		return errFullNodeInvalidFormat
	}

	fn.TCPAddress = data[0]
	fn.APIAddress = data[1]
	fn.KeyID = converter.StrToInt64(data[2])

	if fn.PublicKey, err = hex.DecodeString(data[3]); err != nil {
		return err
	}

	if err = fn.Validate(); err != nil {
		return err
	}

	return nil
}

func (fn *FullNode) Validate() error {
	if fn.KeyID == 0 || len(fn.PublicKey) != publicKeyLength || len(fn.TCPAddress) == 0 {
		return errFullNodeInvalidValues
	}

	if err := ValidateURL(fn.APIAddress); err != nil {
		return err
	}

	return nil
}

// ValidateURL returns error if the URL is invalid
func ValidateURL(rawurl string) error {
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
