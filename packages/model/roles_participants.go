package model

import (
	"fmt"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/converter"
)

type RolesParticipants struct {
	prefix      int64
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
	if prefix == 0 {
		prefix = 1
	}
	r.prefix = prefix
	return r
}

// TableName returns name of table
func (r RolesParticipants) TableName() string {
	if r.prefix == 0 {
		r.prefix = 1
	}
	return fmt.Sprintf("%d_roles_participants", r.prefix)
}

func (r *RolesParticipants) GetActiveMemberRoles(memberID int64) ([]RolesParticipants, error) {
	roles := new([]RolesParticipants)
	err := DBConn.Table(r.TableName()).Where("member->>'member_id' = ? AND deleted = ?", converter.Int64ToStr(memberID), 0).Find(&roles).Error
	return *roles, err
}

// MemberHasRole returns true if member has role
func MemberHasRole(tx *DbTransaction, ecosys, member, role int64) (bool, error) {
	db := GetDB(tx)
	var count int64
	if err := db.Table(fmt.Sprint(ecosys, "_roles_participants")).Where(`role->>'id' = ? and member->>'member_id' = ?`, converter.Int64ToStr(role), converter.Int64ToStr(member)).Count(&count).Error; err != nil {
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
	err = db.Table(fmt.Sprint(ecosys, "_roles_participants")).
		Select("role->>'id'", "role->>'name'").
		Where("member->>'member_id' = ?", converter.Int64ToStr(member)).Find(&ra).Error

	if err != nil {
		return
	}

	for _, role := range ra {
		roles[converter.StrToInt64(role.RoleID)] = role.RoleName
	}

	return
}
