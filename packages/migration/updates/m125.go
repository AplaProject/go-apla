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

var M125 = `ALTER TABLE "1_keys" ADD COLUMN IF NOT EXISTS "address" varchar(255) NOT NULL DEFAULT '';
ALTER TABLE "1_binaries" ADD COLUMN IF NOT EXISTS "address" varchar(255) NOT NULL DEFAULT '';
ALTER TABLE "1_delayed_contracts" ADD COLUMN IF NOT EXISTS "address" varchar(255) NOT NULL DEFAULT '';
ALTER TABLE "1_history" ADD COLUMN IF NOT EXISTS "sender_address" varchar(255) NOT NULL DEFAULT '';
ALTER TABLE "1_history" ADD COLUMN IF NOT EXISTS "recipient_address" varchar(255) NOT NULL DEFAULT '';
ALTER TABLE "1_members" ADD COLUMN IF NOT EXISTS "address" varchar(255) NOT NULL DEFAULT '';
ALTER TABLE "1_buffer_data" ADD COLUMN IF NOT EXISTS "address" varchar(255) NOT NULL DEFAULT '';

UPDATE "1_tables" SET columns=jsonb_set(columns, '{address}', '"ContractConditions(\"@1AdminCondition\")"') WHERE name='keys';
UPDATE "1_tables" SET columns=jsonb_set(columns, '{address}', '"ContractConditions(\"@1AdminCondition\")"') WHERE name='binaries';
UPDATE "1_tables" SET columns=jsonb_set(columns, '{address}', '"ContractAccess(\"@1EditDelayedContract\")"') WHERE name='delayed_contracts';
UPDATE "1_tables" SET columns=jsonb_set(columns, '{sender_address}', '"false"') WHERE name='history';
UPDATE "1_tables" SET columns=jsonb_set(columns, '{recipient_address}', '"false"') WHERE name='history';
UPDATE "1_tables" SET columns=jsonb_set(columns, '{address}', '"false"') WHERE name='members';
UPDATE "1_tables" SET columns=jsonb_set(columns, '{address}', '"false"') WHERE name='buffer_data';

UPDATE "1_tables" SET columns=jsonb_set(columns, '{multi}', '"ContractConditions(\"@1AdminCondition\")"') WHERE name='keys';
`
