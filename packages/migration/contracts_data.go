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

package migration

var contractsDataSQL = `
INSERT INTO "1_contracts" (id, name, value, conditions, app_id, ecosystem)
VALUES
	(next_id('1_contracts'), 'AdminCondition', '// This contract is used to set "admin" rights.
// Usually the "admin" role is used for this.
// The role ID is written to the ecosystem parameter and can be changed.
// The contract requests the role ID from the ecosystem parameter and the contract checks the rights.

contract AdminCondition {
    conditions {
        if EcosysParam("founder_account") == $key_id {
            return
        }

        var role_id_param string
        role_id_param = EcosysParam("role_admin")
        if Size(role_id_param) == 0 {
            warning "Sorry, you do not have access to this action."
        }

        var role_id int
        role_id = Int(role_id_param)
        
        if !RoleAccess(role_id) {
            warning "Sorry, you do not have access to this action."
        }      
    }
}
', 'ContractConditions("MainCondition")', '%[5]d', '%[1]d'),
	(next_id('1_contracts'), 'DeveloperCondition', '// This contract is used to set "developer" rights.
// Usually the "developer" role is used for this.
// The role ID is written to the ecosystem parameter and can be changed.
// The contract requests the role ID from the ecosystem parameter and the contract checks the rights.

contract DeveloperCondition {
	conditions {
		if EcosysParam("founder_account") == $key_id {
            return
        }

        var role_id_param string
        role_id_param = EcosysParam("role_developer")
        if Size(role_id_param) == 0 {
            warning "Sorry, you do not have access to this action."
        }

        var role_id int
        role_id = Int(role_id_param)
        
        if !RoleAccess(role_id) {
            warning "Sorry, you do not have access to this action."
        }      
	}
}
', 'ContractConditions("MainCondition")', '%[5]d', '%[1]d'),
	(next_id('1_contracts'), 'MainCondition', 'contract MainCondition {
	conditions {
		if EcosysParam("founder_account")!=$key_id
		{
			warning "Sorry, you do not have access to this action."
		}
	}
}
', 'ContractConditions("MainCondition")', '%[5]d', '%[1]d');
`
