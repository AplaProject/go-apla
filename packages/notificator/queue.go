// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package notificator

import (
	"github.com/AplaProject/go-apla/packages/types"
)

type Queue struct {
	Accounts []*Accounts
	Roles    []*Roles
}

type Accounts struct {
	Ecosystem int64
	List      []string
}

type Roles struct {
	Ecosystem int64
	List      []int64
}

func (q *Queue) Size() int {
	return len(q.Accounts) + len(q.Roles)
}

func (q *Queue) AddAccounts(ecosystem int64, list ...string) {
	q.Accounts = append(q.Accounts, &Accounts{
		Ecosystem: ecosystem,
		List:      list,
	})
}

func (q *Queue) AddRoles(ecosystem int64, list ...int64) {
	q.Roles = append(q.Roles, &Roles{
		Ecosystem: ecosystem,
		List:      list,
	})
}

func (q *Queue) Send() {
	for _, a := range q.Accounts {
		UpdateNotifications(a.Ecosystem, a.List)
	}

	for _, r := range q.Roles {
		UpdateRolesNotifications(r.Ecosystem, r.List)
	}
}

func NewQueue() types.Notifications {
	return &Queue{
		Accounts: make([]*Accounts, 0),
		Roles:    make([]*Roles, 0),
	}
}
