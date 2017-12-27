package model

import (
	"fmt"
	"testing"
	"time"
)

func TestLoadKeys(t *testing.T) {
	err := GormInit("localhost", 5432, "postgres", "postgres", "apla")
	if err != nil {
		t.Error("can't init gorm")
	}

	err = DBConn.Exec(`truncate table "1_keys";`).Error
	if err != nil {
		t.Error("can't truncate table")
	}

	k := &Key{}
	testKey := k.SetTablePrefix(1)
	testKey.Amount = "1000"
	testKey.PublicKey = []byte("test")
	testKey.ID = 123
	err = DBConn.Save(k).Error
	if err != nil {
		t.Error("error saving test key")
	}

	BufKeys := NewBufferedKeys()
	err = BufKeys.Initialize()
	if err != nil {
		t.Error("can't initialize keys buffer")
	}

	k2 := &Key{}
	testKey2 := k2.SetTablePrefix(1)
	testKey2.Amount = "1000"
	testKey2.PublicKey = []byte("test")
	testKey2.ID = 312
	err = DBConn.Save(k2).Error
	if err != nil {
		t.Error("error saving test key")
	}

	key, found, err := BufKeys.GetKey(1, testKey.ID)
	if err != nil {
		t.Error("error getting key: ", err)
	}

	if !found {
		t.Error("key not found")
	}

	if key.Amount != testKey.Amount {
		t.Errorf("Amount error. Expected: %s, actual: %s", testKey.Amount, key.Amount)
	}

	key2, found, err := BufKeys.GetKey(1, testKey2.ID)
	if err != nil {
		t.Errorf("error getting key2: ", err)
	}

	if !found {
		t.Error("key not found")
	}

	if key2.Amount != testKey2.Amount {
		t.Errorf("Amount 2 error. Expected: %s, actual: %s", testKey2.Amount, key2.Amount)
	}
}

func TestTime(t *testing.T) {
	err := GormInit("localhost", 5432, "postgres", "postgres", "apla")
	if err != nil {
		t.Error("can't init gorm")
	}

	err = DBConn.Exec(`truncate table "1_keys";`).Error
	if err != nil {
		t.Error("can't truncate table")
	}

	k := &Key{}
	testKey := k.SetTablePrefix(1)
	testKey.Amount = "1000"
	testKey.PublicKey = []byte("test")
	testKey.ID = 123
	err = DBConn.Save(k).Error
	if err != nil {
		t.Error("error saving test key")
	}

	startTime := time.Now()
	for i := 0; i < 1000; i++ {
		err := k.Get(123)
		if err != nil {
			t.Errorf("gorm error ", err)
		}
	}
	endTime := time.Now()
	gormTime := endTime.Sub(startTime)

	startTime = time.Now()
	for i := 0; i < 1000; i++ {
		key, found, err := BufKeys.GetKey(1, testKey.ID)
		if err != nil {
			t.Error("error getting key: ", err)
		}

		if !found {
			t.Error("key not found")
		}
		if key.Amount == "" {
			t.Errorf("error amount, ", key.Amount)
		}
	}
	endTime = time.Now()
	bufTime := endTime.Sub(startTime)
	t.Errorf(fmt.Sprintf("gormTime: %s, bufTime: %s", gormTime.String(), bufTime.String()))
}
аьею=