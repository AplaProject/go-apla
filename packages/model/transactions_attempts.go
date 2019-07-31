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

package model

// TransactionsAttempts is model
type TransactionsAttempts struct {
	Hash    []byte `gorm:"primary_key;not null"`
	Attempt int8   `gorm:"not null"`
}

// GetByHash returns TransactionsAttempts existence by hash
func (ta *TransactionsAttempts) GetByHash() (bool, error) {
	return isFound(DBConn.Where("hash = ?", ta.Hash).First(ta))
}

// IncrementTxAttemptCount increases attempt column
func IncrementTxAttemptCount(transactionHash []byte) (int64, error) {
	ta := &TransactionsAttempts{
		Hash: transactionHash,
	}

	found, err := ta.GetByHash()
	if err != nil {
		return 0, err
	}
	if found {
		err = DBConn.Exec("update transactions_attempts set attempt=attempt+1 where hash = ?",
			transactionHash).Error
		if err != nil {
			return 0, err
		}
		ta.Attempt++
	} else {
		ta.Hash = transactionHash
		ta.Attempt = 1
		if err = DBConn.Create(ta).Error; err != nil {
			return 0, err
		}
	}
	return int64(ta.Attempt), nil
}

func DecrementTxAttemptCount(transactionHash []byte) error {
	return DBConn.Exec("update transactions_attempts set attempt=attempt-1 where hash = ?",
		transactionHash).Error
}
