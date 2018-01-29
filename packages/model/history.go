package model

import (
	"errors"
	"fmt"
)

const keysUpdatePattern = `UPDATE "1_keys" SET amount = amount + ? WHERE id = ?`

var errLowBalance = errors.New("not enough funds on the balance")

// History represent record of history table
type History struct {
	ID          int64
	SenderID    int64
	RecipientID int64
	Amount      float64
	Comment     string
	BlockID     int64
	TxHash      []byte
}

// TokenTransferWithHistory add one history record and update wallet balance
func TokenTransferWithHistory(transaction *DbTransaction, history History) error {
	db := GetDB(transaction)
	r := db.Raw(`SELECT amount FROM "1_keys" WHERE id = ?`, history.SenderID).Row()

	var currentSenderAmount float64
	if err := r.Scan(&currentSenderAmount); err != nil {
		return err
	}

	if currentSenderAmount < history.Amount {
		return errLowBalance
	}

	if err := db.Exec(keysUpdatePattern, history.SenderID, -history.Amount).Error; err != nil {
		return fmt.Errorf("sender update amount error: %v", err)
	}

	if err := db.Exec(keysUpdatePattern, history.RecipientID, history.Amount).Error; err != nil {
		return fmt.Errorf("recipient update amount error: %v", err)
	}

	return db.Exec(`INSERT INTO "1_history" (sender_id, recipient_id, amount, comment, block_id, txhash) VALUES (?, ?, ?, ?, ?, ?)`,
		history.SenderID, history.RecipientID, history.Amount, history.Comment, history.BlockID, history.TxHash).Error
}
