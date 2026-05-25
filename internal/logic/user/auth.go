package user

import (
	"SamaraAI/common/code"
	myemail "SamaraAI/common/email"
	myredis "SamaraAI/common/redis"
	"SamaraAI/dao/user"
	"SamaraAI/model"
	"SamaraAI/utils"
	"SamaraAI/utils/myjwt"
	"strings"
)

func login(account, password string) (string, code.Code) {
	account = strings.TrimSpace(account)
	ok, userInformation := resolveUser(account)
	if !ok {
		return "", code.CodeUserNotExist
	}
	if userInformation.Password != password {
		return "", code.CodeInvalidPassword
	}
	token, err := myjwt.GenerateToken(userInformation.ID, userInformation.Username)
	if err != nil {
		return "", code.CodeServerBusy
	}
	return token, code.CodeSuccess
}

func resolveUser(account string) (bool, *model.User) {
	if strings.Contains(account, "@") {
		u, err := user.GetUserByEmail(account)
		if err != nil {
			return false, nil
		}
		return true, u
	}
	return user.IsExistUser(account)
}


func register(email, password, captcha string) (string, code.Code) {
	if u, err := user.GetUserByEmail(email); err == nil && u != nil {
		return "", code.CodeUserExist
	}
	if ok, _ := myredis.CheckCaptchaForEmail(email, captcha); !ok {
		return "", code.CodeInvalidCaptcha
	}

	username := utils.GetRandomNumbers(11)
	userInformation, ok := user.Register(username, email, password)
	if !ok {
		return "", code.CodeServerBusy
	}

	if err := myemail.SendCaptcha(email, username, user.UserNameMsg); err != nil {
		return "", code.CodeServerBusy
	}

	token, err := myjwt.GenerateToken(userInformation.ID, userInformation.Username)
	if err != nil {
		return "", code.CodeServerBusy
	}
	return token, code.CodeSuccess
}

func sendCaptcha(email string) code.Code {
	sendCode := utils.GetRandomNumbers(6)
	if err := myredis.SetCaptchaForEmail(email, sendCode); err != nil {
		return code.CodeServerBusy
	}
	if err := myemail.SendCaptcha(email, sendCode, myemail.CodeMsg); err != nil {
		return code.CodeServerBusy
	}
	return code.CodeSuccess
}
