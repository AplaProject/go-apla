package model

import "github.com/GenesisKernel/go-genesis/packages/consts"

// This constants contains values of transactions priority
const (
	TransactionRateOnBlock transactionRate = iota + 1
	TransactionRateStopNetwork
)

type transactionRate int8

func getTxRateByTxType(txType int8) transactionRate {
	switch txType {
	case consts.TxTypeStopNetwork:
		return TransactionRateStopNetwork
	default:
		return 0
	}
}
