package db

import (
	"log"
	"github.com/jinzhu/gorm"
	"MockTwitter_V2/srv"
)

var dbConn *gorm.DB

//database configurations
func init() {
	if srv.UserDB == nil {
		log.Fatal("Database init failed")
	}
	if err := srv.UserDB.Connect(); err != nil {
		log.Fatal(err)
	}

	dbConn = srv.UserDB.DB()
	if db := dbConn.AutoMigrate(&User{}); db.Error != nil {
		log.Fatal(db.Error)
	}
	if db := dbConn.AutoMigrate(&Post{}); db.Error != nil {
		log.Fatal(db.Error)
	}
}


// Find User according to user's phone number. return nil if user not found
func FindUserByPhoneNumber(phoneNum string) (*User, error) {
	user := User{}
	db := dbConn.First(&user, "phone = ?", phoneNum)
	if db.RecordNotFound() {
		return nil, srv.ErrDBUserNotFound
	}
	if db.Error != nil {
		return nil, db.Error
	}
	return &user, nil
}

// Find User according to user's email. return nil if user not found
func FindUserByEmail(email string) (*User, error) {
	user := User{}
	db := dbConn.First(&user, "email = ?", email)
	if db.RecordNotFound() {
		return nil, srv.ErrDBUserNotFound
	}
	if db.Error != nil {
		return nil, db.Error
	}
	return &user, nil
}

// Find user's phone verification record
func FindPhoneVerification(user *User) (*Post, error) {
	if user == nil {
		return nil, srv.ErrInvalidArg
	}
	ver := make([]Post, 0)
	association := dbConn.Model(user).Association("Verifications")
	if err := association.Error; err != nil {
		return nil, err
	}
	if association.Count() == 0 {
		return nil, srv.ErrDBVerificationNotFound
	}
	if err := association.Find(&ver).Error; err != nil {
		return nil, err
	}
	for _, v := range ver {
		if v.ID == user.Phone {
			return &v, nil
		}
	}
	return nil, srv.ErrDBVerificationNotFound
}

// Find user's email verification record
func FindEmailVerification(user *User) (*Post, error) {
	if user == nil {
		return nil, srv.ErrInvalidArg
	}
	ver := make([]Post, 0)
	association := dbConn.Model(user).Association("Verifications")
	if err := association.Error; err != nil {
		return nil, err
	}
	if association.Count() == 0 {
		return nil, srv.ErrDBVerificationNotFound
	}
	if err := association.Find(&ver).Error; err != nil {
		return nil, err
	}
	for _, v := range ver {
		if v.ID == user.Email {
			return &v, nil
		}
	}
	return nil, srv.ErrDBVerificationNotFound
}

// Create an user record in database and return userID if success
func CreateUserRecord(user *User) (uint, error) {
	if user == nil {
		return 0, srv.ErrInvalidArg
	}
	db := dbConn.Create(user)
	if err := db.Error; err != nil {
		return 0, err
	}
	if err := db.Save(user).Error; err != nil {
		return 0, err
	}
	return user.ID, nil
}

func CreateVerificationRecord(verification *Post) error {
	if verification == nil {
		return srv.ErrInvalidArg
	}
	return dbConn.Create(verification).Error
}

func UpdateVerification(verification *Post) error {
	if verification == nil {
		return srv.ErrInvalidArg
	}
	if verification.ID == "" || verification.UserId == 0 {
		return srv.ErrInvalidArg
	}
	err := dbConn.Model(&Verification{
		ID:     verification.ID,
		UserId: verification.UserId,
	}).Updates(Verification{
		Code:      verification.Code,
		ExpiresAt: verification.ExpiresAt,
	}).Error
	return err
}

func UpdateUserInformation(userid uint, user *User) error {
	if user == nil {
		return srv.ErrInvalidArg
	}
	dbConn.Model(&User{}).Where("id = ?", userid).Updates(user)
	return nil
}

// Check if user has registered before. The user is considered not registered if password is not set.
func IsUserExist(user *User) bool {
	if user != nil {
		if len(user.Password) != 0 {
			return true
		}
	}
	return false
}

// Check if user exists and return user id if true. The user is considered not registered if password is not set.
func CheckUserExistByPhoneNumber(phoneNum string) (uint, bool) {
	user, err := FindUserByPhoneNumber(phoneNum)
	if err == nil && user != nil {
		if len(user.Password) == 0 {
			return 0, false
		} else {
			return user.ID, true
		}
	}
	return 0, false
}
