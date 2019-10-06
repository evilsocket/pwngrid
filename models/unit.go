package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
)

type Unit struct {
	ID          uint32    `gorm:"primary_key; auto_increment" json:"id"`
	Address     string    `gorm:"size:50;not null" json:"address"`
	Country     string    `gorm:"size:10" json:"country"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Fingerprint string    `gorm:"size:255;not null;unique" json:"identity"`
	PublicKey   string    `gorm:"size:10000;not null" json:"public_key"`
	Token       string    `gorm:"size:10000;not null" json:"token"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func FindUnitByFingerprint(db *gorm.DB, fingerprint string) *Unit {
	var unit Unit
	if err := db.Model(&Unit{}).Where("fingerprint = ?", fingerprint).Take(&unit).Error; err != nil {
		return nil
	}
	return &unit
}

func (u Unit) Identity() string {
	return fmt.Sprintf("%s@%s", u.Name, u.Fingerprint)
}
