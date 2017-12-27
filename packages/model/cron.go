package model

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
