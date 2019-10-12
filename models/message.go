package models

import (
	"fmt"
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
	SenderName string     `gorm:"size:255" json:"sender_name"`
	Sender     string     `gorm:"size:255;not null" json:"sender"`
	Data       string     `gorm:"size:512000;not null" json:"-"`
	Signature  string     `gorm:"size:10000;not null" json:"-"`
}

func ValidateMessage(data, signature string) error {
	// validate max sizes
	if dataSize := len(data); dataSize > MessageDataMaxSize {
		return fmt.Errorf("max message data size is %d", MessageDataMaxSize)
	} else if sigSize := len(signature); sigSize > MessageSignatureMaxSize {
		return fmt.Errorf("max message signature size is %d", MessageSignatureMaxSize)
	}
	return nil
}
