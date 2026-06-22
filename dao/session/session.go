package session

import (
	"SamaraAI/common/mysql"
	"SamaraAI/model"
)

func GetSessionsByUserName(userName string) ([]model.Session, error) {
	var sessions []model.Session
	err := mysql.DB.Where("user_name = ?", userName).Order("updated_at desc").Find(&sessions).Error
	return sessions, err
}

func CreateSession(session *model.Session) (*model.Session, error) {
	err := mysql.DB.Create(session).Error
	return session, err
}

func GetSessionByIDAndUser(sessionID, userName string) (*model.Session, error) {
	var session model.Session
	err := mysql.DB.Where("id = ? AND user_name = ?", sessionID, userName).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func DeleteSession(sessionID, userName string) error {
	return mysql.DB.Where("id = ? AND user_name = ?", sessionID, userName).Delete(&model.Session{}).Error
}
