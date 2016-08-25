package controllers

const NBackup = `backup`

type backupPage struct {
	Data        *CommonPage
}

func init() {
	newPage(NBackup)
}

func (c *Controller) Backup() (string, error) {
	return proceedTemplate( c, NBackup, &backupPage{c.Data})
}
