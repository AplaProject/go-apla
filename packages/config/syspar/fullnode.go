package syspar

import (
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/utils"
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

	if err := utils.ValidateURL(fn.APIAddress); err != nil {
		return err
	}

	return nil
}
