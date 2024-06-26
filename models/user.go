package models

import (
	"erinyes/logs"
	"gorm.io/gorm"
)

type User struct {
	ID       int    `gorm:"primaryKey;column:id"`
	Username string `gorm:"column:username"`
	Password string `gorm:"column:password"`
}

func (User) TableName() string {
	return "user"
}

func (u *User) FindByName(db *gorm.DB, username string) bool {
	err := db.First(u, "username = ?", username).Error
	if err != nil {
		logs.Logger.Errorf("can't find user by name = %s", username)
		return false
	}
	return true
}
