package models

import (
	"time"
)

const (
	MessageDataMaxSize      = 512000
	MessageSignatureMaxSize = 10000
)

type Message struct {
	ID         uint       `gorm:"primary_key" json:"id"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `sql:"index" json:"deleted_at"`
	SeenAt     *time.Time `json:"seen_at" sql:"index"`
	SenderID   uint       `json:"-"`
	ReceiverID uint       `json:"-"`
	Sender     string     `gorm:"size:255;not null" json:"sender"`
	Data       string     `gorm:"size:512000;not null" json:"-"`
	Signature  string     `gorm:"size:10000;not null" json:"-"`
}
