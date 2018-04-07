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

func UpdateUserInformation(userid uint, user *User) error {
	if user == nil {
		return srv.ErrInvalidArg
	}
	dbConn.Model(&User{}).Where("id = ?", userid).Updates(user)
	return nil
}
