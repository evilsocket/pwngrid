package models

import (
	"fmt"
	"github.com/evilsocket/islazy/log"
	"github.com/jinzhu/gorm"
)

var db *gorm.DB

func Setup(DbUser, DbPassword, DbPort, DbHost, DbName string) (err error) {
	log.Info("connecting to %s:%s ...", DbHost, DbPort)
	dbURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", DbUser, DbPassword, DbHost, DbPort, DbName)
	if db, err = gorm.Open("mysql", dbURL); err != nil {
		return
	}
	db.Debug().AutoMigrate(&Unit{}, &AccessPoint{})
	return
}

func Create(v interface{}) *gorm.DB {
	return db.Create(v)
}

func Update(v interface{}) *gorm.DB {
	return db.Update(v)
}
