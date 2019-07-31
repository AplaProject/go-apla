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

package model

import (
	"fmt"
)

// Cron represents record of {prefix}_cron table
type Cron struct {
	tableName string
	ID        int64
	Cron      string
	Contract  string
}

// SetTablePrefix is setting table prefix
func (c *Cron) SetTablePrefix(prefix string) {
	c.tableName = prefix + "_cron"
}

// TableName returns name of table
func (c *Cron) TableName() string {
	return c.tableName
}

// Get is retrieving model from database
func (c *Cron) Get(id int64) (bool, error) {
	return isFound(DBConn.Where("id = ?", id).First(c))
}

// GetAllCronTasks is returning all cron tasks
func (c *Cron) GetAllCronTasks() ([]*Cron, error) {
	var crons []*Cron
	err := DBConn.Table(c.TableName()).Find(&crons).Error
	return crons, err
}

// UID returns unique identifier for cron task
func (c *Cron) UID() string {
	return fmt.Sprintf("%s_%d", c.tableName, c.ID)
}
