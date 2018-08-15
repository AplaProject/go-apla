package custom

import (
	"github.com/GenesisCommunity/go-genesis/packages/utils/tx"
)

// TransactionInterface is parsing transactions
type TransactionInterface interface {
	Init() error
	Validate() error
	Action() error
	Rollback() error
	Header() *tx.Header
}
