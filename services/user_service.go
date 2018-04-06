package services

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	sp "github.com/SparkPost/gosparkpost"
	"github.com/nzgogo/micro/codec"
	"gogoexpress-go-api-user-v1/db"
	"gogoexpress-go-api-user-v1/email_template"
	"gogoexpress-go-api-user-v1/server"
	"io"
	"time"
)

var PhoneCodeDigitalTable = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}

const (
	PhoneCodeDuration = 15 * time.Minute
	PhoneCodeLength   = 6
	EmailCodeDuration = 15 * time.Minute

	LoginWithPhone = "p"
	LoginWithEmail = "e"
)

type Email struct {
	Recipients  []Recipient       `json:"recipients"`
	Description string            `json:"description,omitempty"`
	HTML        string            `json:"html"`
	Text        string            `json:"text,omitempty"`
	Subject     string            `json:"subject,omitempty"`
	From        *sp.From          `json:"from,omitempty"`
	ReplyTo     string            `json:"reply_to,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Attachments []sp.Attachment   `json:"attachments,omitempty"`
}

type Recipient struct {
	Email    string `json:"email"`
	Name     string `json:"name,omitempty"`
	HeaderTo string `json:"header_to,omitempty"`
}

/****************************** Mobile Login and Registration Process ******************************/

// User request mobile verification code. return generated verification code
func ReqVerificationCode(phoneNum string) (string, error) {
	var returnCode string
	user, err := database.FindUserByPhoneNumber(phoneNum)
	if err == server.ErrDBUserNotFound {
		newCode := generateCode(PhoneCodeLength)
		returnCode = newCode
		newExpire := time.Now().Add(PhoneCodeDuration)
		_, err = database.CreateUserRecord(&database.User{
			Phone:           phoneNum,
			IsPhoneVerified: false,
			Verifications: []database.Verification{
				{
					ID:        phoneNum,
					Code:      newCode,
					ExpiresAt: newExpire,
				},
			},
		})
		return returnCode, err
	} else if err != nil {
		return "", err
	}

	verification, err := database.FindPhoneVerification(user)
	if err == nil && verification != nil {
		if verification.ExpiresAt.After(time.Now()) {
			returnCode = verification.Code
			newExpire := time.Now().Add(PhoneCodeDuration)
			err = database.UpdateVerification(
				&database.Verification{
					ID:        phoneNum,
					UserId:    verification.UserId,
					Code:      verification.Code,
					ExpiresAt: newExpire,
				})
		} else {
			newCode := generateCode(PhoneCodeLength)
			returnCode = newCode
			newExpire := time.Now().Add(PhoneCodeDuration)
			err = database.UpdateVerification(
				&database.Verification{
					ID:        phoneNum,
					UserId:    verification.UserId,
					Code:      newCode,
					ExpiresAt: newExpire,
				})
		}

	} else if err == server.ErrDBVerificationNotFound && user != nil {
		newCode := generateCode(PhoneCodeLength)
		returnCode = newCode
		newExpire := time.Now().Add(PhoneCodeDuration)
		err = database.CreateVerificationRecord(
			&database.Verification{
				ID:        phoneNum,
				UserId:    user.ID,
				Code:      newCode,
				ExpiresAt: newExpire,
			})
	} else {
		return "", err
	}

	return returnCode, err
}

// Verify user's received verification code. return status code
func CheckVerificationCode(phoneNum, receivedCode string) error {
	user, err := database.FindUserByPhoneNumber(phoneNum)
	if err != nil {
		return err
	}

	verification, err := database.FindPhoneVerification(user)
	if err != nil {
		return err
	}

	if verification.Code != receivedCode {
		return server.ErrVerificationCodeInvalid
	}
	if verification.ExpiresAt.Before(time.Now()) {
		return server.ErrVerificationCodeExpire
	}
	user.IsPhoneVerified = true
	if err = database.UpdateUserInformation(verification.UserId, user); err != nil {
		return err
	}
	// todo delete verification
	return nil
}

func CheckUserExistByPhoneNumber(phoneNum string) (uint, bool) {
	return database.CheckUserExistByPhoneNumber(phoneNum)
}

// Create an user record. return userid and error
func CreateUser(userInfo map[string]interface{}) (uint, error) {
	username, ok1 := userInfo[server.JK_LOGIN_USERNAME].(string)
	pswd, ok2 := userInfo[server.JK_LOGIN_PASSWORD].(string)
	phoneNum, ok3 := userInfo[server.JK_LOGIN_PHONE].(string)
	if !ok1 || !ok2 || !ok3 {
		return 0, server.ErrInvalidJson
	}

	email, ok := userInfo[server.JK_LOGIN_EMAIL].(string)
	if !ok {
		email = ""
	}
	user, err := database.FindUserByPhoneNumber(phoneNum)
	if err == server.ErrDBUserNotFound {
		userid, err := database.CreateUserRecord(&database.User{
			Nickname: username,
			Password: pswd,
			Email:    email,
			Phone:    phoneNum,
		})
		if err != nil {
			return 0, err
		}
		return userid, nil

	} else if err != nil {
		return 0, err
	}
	if user.Nickname != "" || user.Password != "" {
		return 0, server.ErrUserExist
	}
	user.Nickname = username
	user.Password = pswd
	user.Email = email
	err = database.UpdateUserInformation(user.ID, user)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

/****************************** Password Login Process ******************************/

func CheckLoginCredential(account, password, loginMethod string) (uint, error) {
	if account == "" || password == "" {
		return 0, server.ErrMissingField
	}
	var userid uint
	if loginMethod == LoginWithPhone {
		user, err := database.FindUserByPhoneNumber(account)
		if err != nil {
			if err == server.ErrDBUserNotFound {
				return 0, server.ErrInvalidLogin
			} else {
				return 0, err
			}
		}
		if user == nil {
			return 0, server.ErrInvalidLogin
		}
		if !user.IsPhoneVerified {
			return 0, server.ErrInvalidLogin
		}
		if user.Password != password {
			return 0, server.ErrInvalidLogin
		}
		userid = user.ID
	} else {
		user, err := database.FindUserByEmail(account)
		if err != nil {
			if err == server.ErrDBUserNotFound {
				return 0, server.ErrInvalidLogin
			} else {
				return 0, err
			}
		}
		if user == nil {
			return 0, server.ErrInvalidLogin
		}
		if !user.IsEmailVerified {
			return 0, server.ErrUnverifiedEmail
		}
		if user.Password != password {
			return 0, server.ErrInvalidLogin
		}
		userid = user.ID
	}
	return userid, nil
}

/****************************** User Information Modification ******************************/

func PasswordChange(account, password, newPassword string) {

}

func BindEmail(phoneNum, emailAddr string) (msgBody []byte, err error) {
	var user *database.User
	var hashcode string
	user, err = database.FindUserByPhoneNumber(phoneNum)
	if err != nil {
		return
	}
	if user.IsEmailVerified {
		err = server.ErrRecordExist
		return
	}
	user.Email = emailAddr
	err = database.UpdateUserInformation(user.ID, user)
	verification, err1 := database.FindEmailVerification(user)
	err = err1
	if err == nil && verification != nil {
		if verification.ExpiresAt.After(time.Now()) {
			hashcode = verification.Code
			newExpire := time.Now().Add(EmailCodeDuration)
			err = database.UpdateVerification(
				&database.Verification{
					ID:        emailAddr,
					UserId:    verification.UserId,
					Code:      hashcode,
					ExpiresAt: newExpire,
				})
		} else {
			hashcode = GetMD5Hash(emailAddr)
			newExpire := time.Now().Add(EmailCodeDuration)
			err = database.UpdateVerification(
				&database.Verification{
					ID:        emailAddr,
					UserId:    verification.UserId,
					Code:      hashcode,
					ExpiresAt: newExpire,
				})
		}
	} else if err == server.ErrDBVerificationNotFound && user != nil {
		hashcode = GetMD5Hash(emailAddr)
		newExpire := time.Now().Add(EmailCodeDuration)
		err = database.CreateVerificationRecord(
			&database.Verification{
				ID:        emailAddr,
				UserId:    user.ID,
				Code:      hashcode,
				ExpiresAt: newExpire,
			})
	} else {
		return
	}
	var email *Email
	email, err = generateWelcomeEmail(
		user.Nickname,
		user.Email,
		server.SrvEmailValidationURL+server.JK_LOGIN_EMAIL+"="+emailAddr+"&"+server.JK_ACCESS_TOKEN+"="+hashcode)
	msgBody, err = codec.Marshal(email)
	return
}

func ValidateEmailAddress(emailAddr, token string) error {
	user, err := database.FindUserByEmail(emailAddr)
	if err != nil {
		return err
	}
	verification, err := database.FindEmailVerification(user)
	if err == nil && verification != nil {
		if verification.Code != token {
			err = server.ErrVerificationCodeInvalid
		} else if time.Now().After(verification.ExpiresAt) {
			err = server.ErrVerificationCodeExpire
		} else {
			user.IsEmailVerified = true
			database.UpdateUserInformation(user.ID, user)
		}
	}
	return err
}

/****************************** Utilities ******************************/

func generateCode(max int) string {
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		panic(err)
	}
	for i := 0; i < len(b); i++ {
		b[i] = PhoneCodeDigitalTable[int(b[i])%len(PhoneCodeDigitalTable)]
	}
	return string(b)
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func generateWelcomeEmail(username, emailAddr, link string) (email *Email, err error) {
	var html string
	html, err = email_template.WelcomeEn(username, link)
	if err != nil {
		return
	}
	email = &Email{
		Recipients: []Recipient{
			{
				Email: emailAddr,
				Name:  username,
			},
		},
		HTML:    html,
		Subject: "Validate Email",
	}
	return
}
