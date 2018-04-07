package services

import (
	_ "github.com/SparkPost/gosparkpost"
	_ "github.com/nzgogo/micro/codec"
	"MockTwitter_V2/db"
	"MockTwitter_V2/srv"
)

const (
	LoginWithEmail = "e"
)

/****************************** Login and Registration Process ******************************/

// Create an user record. return userid and error
func CreateUser(userInfo map[string]interface{}) (uint, error) {
	username, ok1 := userInfo[srv.JK_LOGIN_USERNAME].(string)
	pswd, ok2 := userInfo[srv.JK_LOGIN_PASSWORD].(string)
	if !ok1 || !ok2 {
		return 0, srv.ErrInvalidJson
	}
	email, ok := userInfo[srv.JK_LOGIN_EMAIL].(string)
	if !ok {
		email = ""
	}
	user, err := db.FindUserByEmail(email)
	if err == srv.ErrDBUserNotFound {
		userid, err := db.CreateUserRecord(&db.User{
			Name: username,
			Email: email,
			Password: pswd,
		})
		if err != nil {
			return 0, err
		}
		return userid, nil

	} else if err != nil {
		return 0, err
	}
	if user.Name != "" || user.Password != "" {
		return 0, srv.ErrUserExist
	}
	user.Name = username
	user.Password = pswd
	user.Email = email
	err = db.UpdateUserInformation(user.ID, user)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

func CheckLoginCredential(account, password, loginMethod string) (uint, error) {
	if account == "" || password == "" {
		return 0, srv.ErrMissingField
	}
	var userid uint
	if loginMethod == LoginWithEmail {
		user, err := db.FindUserByEmail(account)
		if err != nil {
			if err == srv.ErrDBUserNotFound {
				return 0, srv.ErrInvalidLogin
			} else {
				return 0, err
			}
		}
		if user == nil {
			return 0, srv.ErrInvalidLogin
		}
		if user.Password != password {
			return 0, srv.ErrInvalidLogin
		}
		userid = user.ID
	}
	return userid, nil
}