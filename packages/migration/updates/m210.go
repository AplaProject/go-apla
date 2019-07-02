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

var M210 = `
    insert into "1_system_parameters" (id, name, value, conditions) values
 (next_id('1_system_parameters'), 'external_blockchain', '', 'ContractAccess("@1UpdateSysParam")');

    DROP TABLE IF EXISTS "external_blockchain";
	CREATE TABLE "external_blockchain" (
	"id" bigint NOT NULL DEFAULT '0',
	"netname" varchar(255)  NOT NULL DEFAULT '',
	"value" text NOT NULL DEFAULT ''
	);
	ALTER TABLE ONLY "external_blockchain" ADD CONSTRAINT "external_blockchain_pkey" PRIMARY KEY (id);
	CREATE INDEX "external_blockchain_index_name" ON "external_blockchain" (netname);
	
`
