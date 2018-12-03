// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.
//
// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.
//
// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

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
