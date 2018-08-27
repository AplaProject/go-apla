package model

import (
	"time"

	"github.com/GenesisKernel/go-genesis/packages/converter"
)

// RolesParticipants represents record of {prefix}roles_participants table
type RolesParticipants struct {
	ecosystem   int64
	Id          int64
	Role        string `gorm:"type":jsonb(PostgreSQL)`
	Member      string `gorm:"type":jsonb(PostgreSQL)`
	Appointed   string `gorm:"type":jsonb(PostgreSQL)`
	DateCreated time.Time
	DateDeleted time.Time
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
func (r *RolesParticipants) GetActiveMemberRoles(memberID int64) ([]RolesParticipants, error) {
	roles := new([]RolesParticipants)
	err := DBConn.Table(r.TableName()).Where("ecosystem=? and member->>'member_id' = ? AND deleted = ?",
		r.ecosystem, converter.Int64ToStr(memberID), 0).Find(&roles).Error
	return *roles, err
}

// MemberHasRole returns true if member has role
func MemberHasRole(tx *DbTransaction, ecosys, member, role int64) (bool, error) {
	db := GetDB(tx)
	var count int64
	if err := db.Table("1_roles_participants").Where(`ecosystem=? and role->>'id' = ? and member->>'member_id' = ?`,
		ecosys, converter.Int64ToStr(role), converter.Int64ToStr(member)).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetMemberRoles return map[id]name all roles assign to member in ecosystem
func GetMemberRoles(tx *DbTransaction, ecosys, member int64) (roles map[int64]string, err error) {
	db := GetDB(tx)

	var ra []struct {
		RoleID   string
		RoleName string
	}
	err = db.Table("1_roles_participants").
		Select("role->>'id'", "role->>'name'").
		Where("ecosystem=? and member->>'member_id' = ?", ecosys, converter.Int64ToStr(member)).Find(&ra).Error

	if err != nil {
		return
	}

	for _, role := range ra {
		roles[converter.StrToInt64(role.RoleID)] = role.RoleName
	}

	return
}
