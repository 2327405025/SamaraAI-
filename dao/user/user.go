package user

import (
	"SamaraAI/common/mysql"
	"SamaraAI/model"

	"gorm.io/gorm"
)

const UserNameMsg = "SamaraAI的账号如下，请保留好，后续可以用账号或邮箱登录 "

func IsExistUser(username string) (bool, *model.User) {
	u, err := mysql.GetUserByUsername(username)
	if err == gorm.ErrRecordNotFound || u == nil {
		return false, nil
	}
	return true, u
}

func GetUserByEmail(email string) (*model.User, error) {
	return mysql.GetUserByEmail(email)
}

func Register(username, email, password string) (*model.User, bool) {
	u, err := mysql.InsertUser(&model.User{
		Email:    email,
		Name:     username,
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, false
	}
	return u, true
}
