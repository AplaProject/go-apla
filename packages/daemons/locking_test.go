package daemons

import (
	"testing"

	"database/sql"

	"time"

	"context"

	"github.com/EGaaS/go-egaas-mvp/packages/model"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func createTables(t *testing.T, db *sql.DB) {
	sql := `
	CREATE TABLE "main_lock" (
		"lock_time" integer NOT NULL DEFAULT '0',
		"script_name" string NOT NULL DEFAULT '',
		"info" text NOT NULL DEFAULT '',
		"uniq" integer NOT NULL DEFAULT '0'
	);
	CREATE TABLE "install" (
		"progress" text NOT NULL DEFAULT ''
	);
	`
	var err error
	_, err = db.Exec(sql)
	if err != nil {
		t.Fatalf("%s", err)
	}
}

func TestLock(t *testing.T) {
	db := initGorm(t)
	createTables(t, db.DB())

	ctx, _ := context.WithTimeout(context.Background(), 200*time.Millisecond)
	ok, err := DBLock(ctx, "test")
	if err != nil {
		t.Errorf("lock returned %s", err)
	}
	if !ok {
		t.Errorf("can't lock")
	}

	ok, err = tryLock("test2")
	if err != nil {
		t.Errorf("lock returned %s", err)
	}
	if ok {
		t.Errorf("lock should fail")
	}

	ml := &model.MainLock{}
	err = ml.Get()
	if err != nil {
		t.Fatalf("Get main_lock failed: %s", err)
	}
	if ml.ScriptName != "test" {
		t.Errorf("bad script_name: want test, got %s", ml.ScriptName)
	}

	time.Sleep(1 * time.Second)
	err = UpdMainLock()
	if err != nil {
		t.Fatalf("update main lock failed: %s", err)
	}
	ml2 := &model.MainLock{}
	err = ml2.Get()
	if err != nil {
		t.Fatalf("Get main_lock failed: %s", err)
	}
	if ml2.ScriptName != "test" {
		t.Errorf("bad script_name: want test, got %s", ml.ScriptName)
	}
	if ml2.LockTime == ml.LockTime {
		t.Errorf("UpdMainLock didn't change the lock time")
	}

}

func TestUnlock(t *testing.T) {
	db := initGorm(t)
	createTables(t, db.DB())

	ok, err := tryLock("test")
	if err != nil {
		t.Errorf("lock returned %s", err)
	}

	if !ok {
		t.Errorf("can't lock")
	}

	// try another goroutine name
	err = DBUnlock("some_another_name")
	if err != nil {
		t.Errorf("DBUnlock error: %s", err)
	}

	ok, err = tryLock("some_another_name")
	if err != nil {
		t.Errorf("lock returned %s", err)
	}

	if ok {
		t.Errorf("incorrect lock")
	}

	// try unlock
	err = DBUnlock("test")
	if err != nil {
		t.Errorf("DBUnlock error: %s", err)
	}

	ok, err = tryLock("some_another_name")
	if err != nil {
		t.Errorf("lock returned %s", err)
	}

	if !ok {
		t.Errorf("lock failed")
	}

}

func TestWait(t *testing.T) {
	db := initGorm(t)
	createTables(t, db.DB())

	ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)
	err := WaitDB(ctx)
	if err == nil {
		t.Errorf("should be error")
	}

	install := &model.Install{}
	install.Progress = "complete"
	err = install.Save()
	if err != nil {
		t.Fatalf("save failed: %s", err)
	}

	ctx, _ = context.WithTimeout(context.Background(), 100*time.Millisecond)
	err = WaitDB(ctx)
	if err != nil {
		t.Errorf("wait failed: %s", err)
	}
}
