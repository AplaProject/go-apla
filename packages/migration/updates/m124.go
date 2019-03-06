// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

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

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package updates

var M124 = `
ALTER TABLE "1_notifications" ADD COLUMN "temp" bigint NOT NULL DEFAULT '0';
UPDATE "1_notifications" SET temp = EXTRACT(EPOCH FROM date_created::timestamp with time zone) WHERE date_created IS NOT NULL;
ALTER TABLE "1_notifications" DROP COLUMN "date_created";
ALTER TABLE "1_notifications" RENAME COLUMN "temp" TO "date_created";

ALTER TABLE "1_notifications" ADD COLUMN "temp" bigint NOT NULL DEFAULT '0';
UPDATE "1_notifications" SET temp = EXTRACT(EPOCH FROM date_start_processing::timestamp with time zone) WHERE date_start_processing IS NOT NULL;
ALTER TABLE "1_notifications" DROP COLUMN "date_start_processing";
ALTER TABLE "1_notifications" RENAME COLUMN "temp" TO "date_start_processing";

ALTER TABLE "1_notifications" ADD COLUMN "temp" bigint NOT NULL DEFAULT '0';
UPDATE "1_notifications" SET temp = EXTRACT(EPOCH FROM date_closed::timestamp with time zone) WHERE date_closed IS NOT NULL;
ALTER TABLE "1_notifications" DROP COLUMN "date_closed";
ALTER TABLE "1_notifications" RENAME COLUMN "temp" TO "date_closed";

ALTER TABLE "1_roles" ADD COLUMN "temp" bigint NOT NULL DEFAULT '0';
UPDATE "1_roles" SET temp = EXTRACT(EPOCH FROM date_created::timestamp with time zone) WHERE date_created IS NOT NULL;
ALTER TABLE "1_roles" DROP COLUMN "date_created";
ALTER TABLE "1_roles" RENAME COLUMN "temp" TO "date_created";

ALTER TABLE "1_roles" ADD COLUMN "temp" bigint NOT NULL DEFAULT '0';
UPDATE "1_roles" SET temp = EXTRACT(EPOCH FROM date_deleted::timestamp with time zone) WHERE date_deleted IS NOT NULL;
ALTER TABLE "1_roles" DROP COLUMN "date_deleted";
ALTER TABLE "1_roles" RENAME COLUMN "temp" TO "date_deleted";

ALTER TABLE "1_roles_participants" ADD COLUMN "temp" bigint NOT NULL DEFAULT '0';
UPDATE "1_roles_participants" SET temp = EXTRACT(EPOCH FROM date_created::timestamp with time zone) WHERE date_created IS NOT NULL;
ALTER TABLE "1_roles_participants" DROP COLUMN "date_created";
ALTER TABLE "1_roles_participants" RENAME COLUMN "temp" TO "date_created";

ALTER TABLE "1_roles_participants" ADD COLUMN "temp" bigint NOT NULL DEFAULT '0';
UPDATE "1_roles_participants" SET temp = EXTRACT(EPOCH FROM date_deleted::timestamp with time zone) WHERE date_deleted IS NOT NULL;
ALTER TABLE "1_roles_participants" DROP COLUMN "date_deleted";
ALTER TABLE "1_roles_participants" RENAME COLUMN "temp" TO "date_deleted";
`
