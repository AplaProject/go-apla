package controllers

const NBackup = `backup`

type backupPage struct {
	CommonPage
	Address     string
}

func init() {
	newPage(NBackup)
}

func (c *Controller) Backup() (string, error) {
	return proceedTemplate( c, NBackup, &backupPage{Address: c.SessAddress})
}
