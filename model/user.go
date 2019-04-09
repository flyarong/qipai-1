package model

import (
	"github.com/jinzhu/gorm"
	"qipai/enum"
)



type Auth struct {
	gorm.Model
	UserType enum.UserType `gorm:"type:int;not null"`
	Name     string        `gorm:"size:50"`
	Pass     string        `gorm:"size:50"`
	Verified bool
	UserId   int
	User     User
}

type User struct {
	gorm.Model
	Nick    string `gorm:"size:20"`
	Mobile  string `gorm:"size:20"`
	Ip      string `gorm:"size:20"`
	Address string `gorm:"size:50"`
	Auths   []Auth
}
