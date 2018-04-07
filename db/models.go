package db

import "github.com/jinzhu/gorm"


//User database
type User struct {
	gorm.Model
	Name      string	`json:"name"`
	Email     string	`json:"email"`
	Password  string	`json:"password"`
}

//Post database
type Post struct {
	gorm.Model
	Post      string	`json:"post"`
	User      User 		`gorm:"foreignkey:UserRefer"` // use UserRefer as foreign key
	UserRefer string
}