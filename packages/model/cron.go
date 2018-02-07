// MIT License
//
// Copyright (c) 2016 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package model

import (
	"fmt"
)

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
