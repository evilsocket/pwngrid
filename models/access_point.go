package models

import "github.com/jinzhu/gorm"

type AccessPoint struct {
	gorm.Model

	UnitID uint   `json:"-"`
	Name  string `gorm:"size:255;not null" json:"name"`
	Mac  string `gorm:"size:255;not null" json:"mac"`
}
