package user

import (
	"fmt"
	"errors"
	"github.com/nzgogo/micro/codec"
	"github.com/nzgogo/micro/router"
	"MockTwitter_V2/srv"
	"MockTwitter_V2/services"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var dbConn *gorm.DB

//message body json value
const (
	// code validation key and values
	USER_NOT_EXSIT = "USER_NOT_EXIST"
)

/****************************** Login and Registration Process ******************************/

// Return all the posts
func Show(msg *codec.Message, reply string) *router.Error {
	userid, err := services.CreateUser(msg.GetAll())
	if err != nil {
		return srv.ErrorCode(err)
	}

	// create user success. request access token from auth server.
	body, err := requestAccessToken(msg, userid)
	if err != nil {
		return srv.ErrorCode(err)
	}

	// pack response message
	resp := codec.NewJsonResponse(msg.ContextID, 201, body)
	err = srv.Service.Respond(resp, reply)
	if err != nil {
		return srv.ErrorCode(err)
	}
	return nil
}

// Create an user record. message should contain at least "Username", "Email", "Password"
func CreateUser(msg *codec.Message, reply string) *router.Error {
	userid, err := services.CreateUser(msg.GetAll())
	if err != nil {
		return srv.ErrorCode(err)
	}

	// create user success. request access token from auth server.
	body, err := requestAccessToken(msg, userid)
	if err != nil {
		return srv.ErrorCode(err)
	}

	// pack response message
	resp := codec.NewJsonResponse(msg.ContextID, 201, body)
	err = srv.Service.Respond(resp, reply)
	if err != nil {
		return srv.ErrorCode(err)
	}
	return nil
}

// Create an Post record. message should contain at least "Postmsg", "UserRefer"
func Post(msg *codec.Message, reply string) *router.Error {
	var phoneNum string
	var vCode string
	phoneNum, ok := msg.Get(srv.JK_LOGIN_PHONE).(string)
	if !ok {
		return srv.ErrorCode(srv.ErrInvalidJson)
	}
	vCode, ok = msg.Get(srv.JK_VERIFY_CODE).(string)
	if !ok {
		return srv.ErrorCode(srv.ErrInvalidJson)
	}
	println(phoneNum, vCode)
	//err := services.CheckVerificationCode(phoneNum, vCode)
	//if err != nil {
	//	return srv.ErrorCode(err)
	//}
	//
	//body := make(map[string]string)
	//var resp *codec.Message
	//
	////if userid, exist := services.CheckUserExistByPhoneNumber(phoneNum); exist {
	//	token, err := requestAccessToken(msg, userid)
	//	if err != nil {
	//		return srv.ErrorCode(err)
	//	}
	//	resp = codec.NewJsonResponse(msg.ContextID, 201, token)
	//} else {
	//	body[srv.JK_MESSAGE] = USER_NOT_EXSIT
	//	resp = codec.NewJsonResponse(msg.ContextID, 201, body)
	//}
	//
	//err = srv.Service.Respond(resp, reply)
	//if err != nil {
	//	return srv.ErrorCode(err)
	//}
	return nil
}

// User Logout
func UserLogOut(msg *codec.Message, reply string) *router.Error {
	if err := revokeToken(msg, msg.Header.Get(srv.HK_ACCESS_TOKEN)); err != nil {
		return srv.ErrorCode(err)
	}
	// pack response message
	resp := codec.NewJsonResponse(msg.ContextID, 200, nil)
	err := srv.Service.Respond(resp, reply)
	if err != nil {
		return srv.ErrorCode(err)
	}
	return nil
}

/****************************** Password Login Process ******************************/

func LoginWithCredential(msg *codec.Message, reply string) *router.Error {
	acc, ok1 := msg.Get(srv.JK_ACCOUNT).(string)
	pswd, ok2 := msg.Get(srv.JK_LOGIN_PASSWORD).(string)
	lm, ok3 := msg.Get(srv.JK_LOGIN_METHOD).(string)
	if !ok1 || !ok2 || !ok3 {
		return srv.ErrorCode(srv.ErrInvalidJson)
	}

	// todo add check-login-method wrapper
	userid, err := services.CheckLoginCredential(acc, pswd, lm)
	if err != nil {
		return srv.ErrorCode(err)
	}

	// Login success, request access token from auth server.
	body, err := requestAccessToken(msg, userid)
	if err != nil {
		return srv.ErrorCode(err)
	}

	// pack response message
	resp := codec.NewJsonResponse(msg.ContextID, 201, body)
	err = srv.Service.Respond(resp, reply)
	if err != nil {
		return srv.ErrorCode(err)
	}
	return nil
}

/****************************** Internal Communication ******************************/

// Request. Send a request to auth service to obtain an access token
func requestAccessToken(msg *codec.Message, userid uint) ([]byte, error) {
	var ret []byte
	subject, err := srv.Service.Options().Selector.Select(srv.SrvAuth, srv.SrvAuthVersion)
	if err != nil {
		return nil, err
	}

	body := make(map[string]interface{})
	body[srv.JK_LOGIN_PHONE] = msg.Get(srv.JK_LOGIN_PHONE)
	body[srv.JK_USER_ID] = userid

	mbody, err := codec.Marshal(body)
	if err != nil {
		return nil, err
	}
	msg.Body = mbody
	msg.Node = srv.SrvAuthIssueAccessTokenHdler
	resp, err := codec.Marshal(msg)
	if err != nil {
		return nil, err
	}
	err = srv.Service.Options().Transport.Request(subject, resp, func(bytes []byte) error {
		reqMsg := &codec.Message{}
		err = codec.Unmarshal(bytes, reqMsg)
		if err != nil {
			return err
		}
		if reqMsg.StatusCode < 200 || reqMsg.StatusCode > 299 {
			errMsg, ok := reqMsg.Get(srv.JK_MESSAGE).(string)
			if !ok {
				return errors.New("")
			}
			return errors.New(errMsg)
		}
		ret = reqMsg.Body
		return nil
	})

	return ret, err
}

// Request. call auth service revoke token
func revokeToken(msg *codec.Message, token string) error {
	subject, err := srv.Service.Options().Selector.Select(srv.SrvAuth, srv.SrvAuthVersion)
	if err != nil {
		return err
	}

	body := make(map[string]string)
	body[srv.JK_ACCESS_TOKEN] = token

	mbody, err := codec.Marshal(body)
	if err != nil {
		return err
	}
	msg.Body = mbody
	msg.Node = srv.SrvAuthRevokeAccessTokenHdler
	resp, err := codec.Marshal(msg)
	if err != nil {
		return err
	}
	err = srv.Service.Options().Transport.Request(subject, resp, func(bytes []byte) error {
		reqMsg := &codec.Message{}
		err = codec.Unmarshal(bytes, reqMsg)
		if err != nil {
			return err
		}
		if reqMsg.StatusCode < 200 || reqMsg.StatusCode > 299 {
			errMsg, ok := reqMsg.Get(srv.JK_MESSAGE).(string)
			if !ok {
				return errors.New("")
			}
			return errors.New(errMsg)
		}
		fmt.Println(reqMsg)
		return nil
	})

	return err
}

