package message

import (
	"SamaraAI/common/mysql"
	"SamaraAI/model"
)

func GetMessagesBySessionID(sessionID string) ([]model.Message, error) {
	var msgs []model.Message
	err := mysql.DB.Where("session_id = ?", sessionID).Order("created_at asc").Find(&msgs).Error
	return msgs, err
}

func CreateMessage(message *model.Message) (*model.Message, error) {
	err := mysql.DB.Create(message).Error
	return message, err
}

func GetAllMessages() ([]model.Message, error) {
	var msgs []model.Message
	err := mysql.DB.Order("created_at asc").Find(&msgs).Error
	return msgs, err
}

func DeleteMessagesBySessionID(sessionID string) error {
	return mysql.DB.Where("session_id = ?", sessionID).Delete(&model.Message{}).Error
}
