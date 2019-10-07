package models

import "github.com/jinzhu/gorm"

type AccessPoint struct {
	gorm.Model

	UnitID uint   `json:"-"`
	ESSID  string `gorm:"size:255;not null" json:"essid"`
	BSSID  string `gorm:"size:255;not null" json:"bssid"`
}
