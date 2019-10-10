package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

const (
	MessageDataMaxSize      = 512000
	MessageSignatureMaxSize = 10000
)

type Message struct {
	gorm.Model

	SeenAt     time.Time `json:"seen_at"`
	SenderID   uint      `json:"-"`
	ReceiverID uint      `json:"-"`
	Data       string    `gorm:"size:512000;not null" json:"data"`
	Signature  string    `gorm:"size:10000;not null" json:"-"`
}
