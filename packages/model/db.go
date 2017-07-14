package model

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	DBConn *gorm.DB
)

func GormInit(user string, pass string, host string, dbName string) error {
	var err error
	DBConn, err = gorm.Open("postgres", fmt.Sprintf(
		"user=%s password=mypassword host=%s dbname=%s sslmode=disable "), user, pass, host, dbName)
	if err != nil {
		return err
	}
	return nil
}
