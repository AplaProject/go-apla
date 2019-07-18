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

package model

import (
	"github.com/AplaProject/go-apla/packages/converter"
)

// RolesParticipants represents record of {prefix}roles_participants table
type RolesParticipants struct {
	ecosystem   int64
	Id          int64
	Role        string `gorm:"type":jsonb(PostgreSQL)`
	Member      string `gorm:"type":jsonb(PostgreSQL)`
	Appointed   string `gorm:"type":jsonb(PostgreSQL)`
	DateCreated int64
	DateDeleted int64
	Deleted     bool
}

// SetTablePrefix is setting table prefix
func (r *RolesParticipants) SetTablePrefix(prefix int64) *RolesParticipants {
	r.ecosystem = prefix
	return r
}

// TableName returns name of table
func (r RolesParticipants) TableName() string {
	if r.ecosystem == 0 {
		r.ecosystem = 1
	}
	return "1_roles_participants"
}

// GetActiveMemberRoles returns active assigned roles for memberID
func (r *RolesParticipants) GetActiveMemberRoles(account string) ([]RolesParticipants, error) {
	roles := new([]RolesParticipants)
	err := DBConn.Table(r.TableName()).Where("ecosystem=? and member->>'account' = ? AND deleted = ?",
		r.ecosystem, account, 0).Find(&roles).Error
	return *roles, err
}

// MemberHasRole returns true if member has role
func MemberHasRole(tx *DbTransaction, role, ecosys int64, account string) (bool, error) {
	db := GetDB(tx)
	var count int64
	if err := db.Table("1_roles_participants").Where(`ecosystem=? and role->>'id' = ? and member->>'account' = ?`,
		ecosys, converter.Int64ToStr(role), account).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetMemberRoles return map[id]name all roles assign to member in ecosystem
func GetMemberRoles(tx *DbTransaction, ecosys int64, account string) (roles []int64, err error) {
	query := `SELECT role->>'id' as "id" 
		FROM "1_roles_participants"
		WHERE ecosystem = ? and deleted = '0' and member->>'account' = ?`
	list, err := GetAllTransaction(tx, query, -1, ecosys, account)
	if err != nil {
		return
	}
	for _, role := range list {
		roles = append(roles, converter.StrToInt64(role[`id`]))
	}
	return
}

// GetRoleMembers return []id all members assign to roles in ecosystem
func GetRoleMembers(tx *DbTransaction, ecosys int64, roles []int64) (members []string, err error) {
	rolesList := make([]string, 0, len(roles))
	for _, role := range roles {
		rolesList = append(rolesList, converter.Int64ToStr(role))
	}
	query := `SELECT member->>'account' as "id" 
		FROM "1_roles_participants" 
		WHERE role->>'id' in (?) group by 1`
	list, err := GetAllTransaction(tx, query, -1, rolesList)
	if err != nil {
		return
	}
	for _, member := range list {
		members = append(members, member[`id`])
	}
	return
}
