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

package vde

var menuDataSQL = `INSERT INTO "1_menu" (id, name, value, conditions, ecosystem) VALUES
(next_id('1_menu'), 'admin_menu', 'MenuItem(Title:"Application", Page:apps_list, Icon:"icon-folder")
MenuItem(Title:"Ecosystem parameters", Page:params_list, Icon:"icon-settings")
MenuItem(Title:"Menu", Page:menus_list, Icon:"icon-list")
MenuItem(Title:"Confirmations", Page:confirmations, Icon:"icon-check")
MenuItem(Title:"Import", Page:import_upload, Icon:"icon-cloud-upload")
MenuItem(Title:"Export", Page:export_resources, Icon:"icon-cloud-download")
MenuGroup(Title:"Resources", Icon:"icon-share"){
	MenuItem(Title:"Pages", Page:app_pages, Icon:"icon-screen-desktop")
	MenuItem(Title:"Blocks", Page:app_blocks, Icon:"icon-grid")
	MenuItem(Title:"Tables", Page:app_tables, Icon:"icon-docs")
	MenuItem(Title:"Contracts", Page:app_contracts, Icon:"icon-briefcase")
	MenuItem(Title:"Application parameters", Page:app_params, Icon:"icon-wrench")
	MenuItem(Title:"Language resources", Page:app_langres, Icon:"icon-globe")
	MenuItem(Title:"Binary data", Page:app_binary, Icon:"icon-layers")
}
MenuItem(Title:"Dashboard", Page:admin_dashboard, Icon:"icon-wrench")', 'ContractConditions("MainCondition")', '%[1]d');
`
