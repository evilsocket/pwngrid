package models

import (
	"fmt"
	"github.com/evilsocket/islazy/log"
	"github.com/jinzhu/gorm"
	"os"
)

var db *gorm.DB

func Setup() (err error) {
	hostname := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	username := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")

	log.Info("connecting to %s:%s ...", hostname, port)
	dbURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, hostname, port, name)
	if db, err = gorm.Open("mysql", dbURL); err != nil {
		return
	}
	db.Debug().AutoMigrate(&Unit{}, &AccessPoint{}, &Message{})
	return
}

func Create(v interface{}) *gorm.DB {
	return db.Create(v)
}

func Update(v interface{}) *gorm.DB {
	return db.Model(v).Update(v)
}

func UpdateFields(v interface{}, fields map[string]interface{}) *gorm.DB {
	return db.Model(v).Updates(fields)
}
