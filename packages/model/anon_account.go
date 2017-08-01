package model

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type AnonAmount struct {
	IDCitizen int64
	IDAnonym  int64
	Encrypted []byte
	Amount    decimal.Decimal
}

func (aa *AnonAmount) Get(tablePrefix int64, citizenID int64) ([]AnonAmount, error) {
	var result []AnonAmount
	err := DBConn.Table(string(tablePrefix) + "_anonyms").
		Select("anon.id_citizen, anon.id_anonym, anon.encrypted, acc.amount").
		Joins(fmt.Sprintf("left join %d_accounts as acc on acc.citizen_id=anon.id_anonym where, anon.id_citizen=?",
			tablePrefix, citizenID)).Scan(result).Error
	return result, err
}
