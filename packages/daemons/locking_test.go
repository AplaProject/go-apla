//MIT License
//
//Copyright (c) 2016-2018 GenesisKernel
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.

package daemons

import (
	"testing"

	"database/sql"

	"time"

	"context"

	"github.com/GenesisKernel/go-genesis/packages/model"
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

func TestWait(t *testing.T) {
	db := initGorm(t)
	createTables(t, db.DB())

	ctx, cf := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer func() {
		ctx.Done()
		cf()
	}()

	err := WaitDB(ctx)
	if err == nil {
		t.Errorf("should be error")
	}

	install := &model.Install{}
	install.Progress = "complete"
	err = install.Create()
	if err != nil {
		t.Fatalf("save failed: %s", err)
	}

	ctx, scf := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer func() {
		ctx.Done()
		scf()
	}()

	err = WaitDB(ctx)
	if err != nil {
		t.Errorf("wait failed: %s", err)
	}
}
