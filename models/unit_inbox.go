package models

import "github.com/biezhi/gorm-paginator/pagination"

func (u *Unit) GetPagedInbox(page int) (messages []Message, total int, pages int) {
	query := db.Model(Message{}).Where("receiver_id = ?", u.ID)
	paginator := pagination.Paging(&pagination.Param{
		DB:      query,
		Page:    page,
		Limit:   50,
		OrderBy: []string{"id desc"},
	}, &messages)
	return messages, paginator.TotalRecord, paginator.TotalPage
}

func (u *Unit) GetInboxMessage(id int) *Message {
	var msg Message
	if err := db.Where("receiver_id = ? AND id = ?", u.ID, id).First(&msg).Error; err != nil{
		return nil
	}
	return &msg
}
