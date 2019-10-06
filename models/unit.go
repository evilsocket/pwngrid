package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
)

type Unit struct {
	ID        uint32    `gorm:"primary_key; auto_increment" json:"id"`
	Address   string    `gorm:"size:50;not null" json:"address"`
	Identity  string    `gorm:"size:255;not null;unique" json:"identity"`
	PublicKey string    `gorm:"size:10000;not null" json:"public_key"`
	Token     string    `gorm:"size:10000;not null" json:"token"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func FindUnitByFingerprint(db *gorm.DB, fingerprint string) *Unit {
	var unit Unit
	search := fmt.Sprintf("%%@%s", fingerprint)
	if err := db.Model(&Unit{}).Where("identity LIKE ?", search).Take(&unit).Error; err != nil {
		return nil
	}
	return &unit
}