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

package daemons

import (
	"testing"

	"database/sql"

	"time"

	"context"

	"github.com/AplaProject/go-apla/packages/model"
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
