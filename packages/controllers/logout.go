package controllers

func (c *Controller) Logout() (string, error) {
	c.sess.Delete("wallet_id")
	c.sess.Delete("citizen_id")
	c.sess.Delete("address")
	return "", nil
}
