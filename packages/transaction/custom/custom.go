package custom

import (
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"
)

// TransactionInterface is parsing transactions
type TransactionInterface interface {
	Init() error
	Validate() error
	Action() error
	Rollback() error
	Header() *tx.Header
}
