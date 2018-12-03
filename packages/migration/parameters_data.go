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

package migration

var parametersDataSQL = `
INSERT INTO "1_parameters" ("id","name", "value", "conditions", "ecosystem") VALUES
	(next_id('1_parameters'),'founder_account', '%[2]d', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'new_table', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'changing_tables', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'changing_language', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'changing_signature', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'changing_page', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'changing_menu', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'changing_contracts', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'max_sum', '1000000', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'stylesheet', 'body {
		  /* You can define your custom styles here or create custom CSS rules */
	}', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'max_tx_block_per_user', '100', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'min_page_validate_count', '1', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'max_page_validate_count', '6', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'changing_blocks', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")', '%[1]d');
`
