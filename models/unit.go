package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Unit struct {
	ID        uint32    `gorm:"primary_key;auto_increment" json:"id"`
	Address   string    `gorm:"size:50;not null" json:"address"`
	Identity  string    `gorm:"size:255;not null;unique" json:"identity"`
	PublicKey string    `gorm:"size:1024;not null;unique" json:"public_key"`
	Token     string    `gorm:"size:255;not null;unique" json:"token"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func FindUnitByIdentity(db *gorm.DB, identity string) (error, *Unit) {
	var unit Unit
	if err := db.Model(&Unit{}).Where("identity = ?", identity).Take(&unit).Error; err != nil {
		return err, nil
	}
	return nil, &unit
}
