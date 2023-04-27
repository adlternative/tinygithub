package model

import "gorm.io/gorm"

type Password struct {
	gorm.Model
	UserID   uint   `gorm:"uniqueIndex"`
	Password string `gorm:"not null"`
}
