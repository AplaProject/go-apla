package model

import (
	"fmt"
	"time"
)

// RolesAssign represent role_asign record
type RolesAssign struct {
	ID              int64
	RoleID          int64
	RoleName        string
	MemberID        int64
	MemberName      string
	MemberAvatar    string
	AppointedByID   int64
	AppointedByName string
	DateStart       time.Time
	DateEnd         time.Time
	Delete          int64
}

// MemberHasRole returns true if member has role
func MemberHasRole(tx *DbTransaction, ecosys, member, role int64) (bool, error) {
	db := GetDB(tx)
	var count int64
	if err := db.Table(fmt.Sprint(ecosys, "_roles_assign")).Where("role_id = ? and member_id = ?", role, member).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetMemberRoles return map[id]name all roles assign to member in ecosystem
func GetMemberRoles(tx *DbTransaction, ecosys, member int64) (roles map[int64]string, err error) {
	db := GetDB(tx)

	var ra []RolesAssign
	err = db.Table(fmt.Sprint(ecosys, "_roles_assign")).
		Select("role_id", "role_name").
		Where("member_id = ?", member).Find(&ra).Error

	if err != nil {
		return
	}

	for _, role := range ra {
		roles[role.RoleID] = role.RoleName
	}

	return
}
