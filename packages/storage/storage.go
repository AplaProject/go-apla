package storage

import (
	"github.com/AplaProject/go-apla/packages/storage/multi"
)

var M *multi.MultiStorage

const (
	typeMem      = "mem"
	typeRegistry = "reg"
)

func NewMultiTransaction() (*multi.MultiTransaction, error) {
	return M.Begin()
}
