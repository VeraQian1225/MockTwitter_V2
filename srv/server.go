package srv

import (
	"errors"
	"github.com/nzgogo/micro"
	"github.com/nzgogo/micro/db"
	"github.com/nzgogo/micro/router"
	"net/http"
)

// Service info and configs
const (
	VERSION = "v1"
	SRVNAME = "gogo-core-user"

	SrvAuth            	= "gogo-core-auth"
	SrvAuthVersion		= "v1"
	SrvAuthIssueAccessTokenHdler  = "get_new_token"
	SrvAuthRevokeAccessTokenHdler = "revoke_token"
	)

// json key/value in transport message
const (
	JK_LOGIN_USERNAME = "username"
	JK_LOGIN_EMAIL    = "email"
	JK_LOGIN_PASSWORD = "password"
	JK_ACCESS_TOKEN   = "token"
	JK_USER_ID        = "userID"
	JK_MESSAGE        = "message"
	JK_ACCOUNT        = "account"
	JK_LOGIN_METHOD   = "loginMethod"
	HK_ACCESS_TOKEN   = "Authorization"
)

// Service and Database instance
var (
	Service = gogo.NewService(SRVNAME, VERSION)
	//UserDB = database.SetupDB(Service.Config())
	UserDB = db.NewDB(
		Service.Config()["db_userDB_username"],
		Service.Config()["db_userDB_password"],
		Service.Config()["db_userDB_name"],
		db.Dialects(Service.Config()["db_userDB_dialect"]),
		db.Address(Service.Config()["db_userDBAddress"]))
)

// errors
var (
	// internal errors
	ErrDBUserNotFound         = errors.New("user not found")
	ErrDBVerificationNotFound = errors.New("verification not found")

	// 4xx
	ErrMissingField            = errors.New("missed some fields")
	ErrUserExist               = errors.New("user already exist")
	ErrNoPhoneNumber           = errors.New("phone not presented in request")
	ErrVerificationCodeInvalid = errors.New("verification code is invalid")
	ErrVerificationCodeExpire  = errors.New("verification code has expired")
	ErrInvalidLogin            = errors.New("invalid username or password")
	ErrUnverifiedEmail         = errors.New("email not verified")
	ErrInvalidJson             = errors.New("json format error")
	ErrRecordExist             = errors.New("record already exists")

	// 5xx
	ErrInvalidArg      = errors.New("invalid argument")
	ErrInvalidResponse = errors.New("internal transport error: invalid response")
)

// error class
func ErrorCode(err error) *router.Error {
	var statusCode int
	switch err {
	case nil:
		return nil
		//4×× Client Error
	case ErrDBUserNotFound:
		statusCode = http.StatusUnprocessableEntity
	case ErrNoPhoneNumber:
		statusCode = http.StatusBadRequest
	case ErrVerificationCodeInvalid:
		statusCode = http.StatusUnprocessableEntity
	case ErrVerificationCodeExpire:
		statusCode = http.StatusUnprocessableEntity
	case ErrMissingField:
		statusCode = http.StatusUnprocessableEntity
	case ErrUserExist:
		statusCode = http.StatusUnprocessableEntity
	case ErrInvalidLogin:
		statusCode = http.StatusUnprocessableEntity
	case ErrUnverifiedEmail:
		statusCode = http.StatusUnprocessableEntity

		// 5xx Server Error
	case ErrInvalidArg:
		statusCode = http.StatusInternalServerError
	case ErrInvalidResponse:
		statusCode = http.StatusInternalServerError
	default:
		statusCode = http.StatusInternalServerError
	}
	return &router.Error{StatusCode: statusCode, Message: err.Error()}

}