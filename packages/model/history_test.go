package model

import (
	"testing"
)

func TestUpdate(t *testing.T) {
	err := GormInit("localhost", 5432, "postgres", "postgres", "apla2")

	if err != nil {
		t.Error(err)
	}

	if err := GetDB(nil).Exec(`UPDATE "1_keys" SET amount = amount + ? WHERE id = ?`, -20000, 2634685165508018480).Error; err != nil {
		t.Error(err)
	}

}
