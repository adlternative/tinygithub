package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Name  string `gorm:"uniqueIndex;not null" form:"username" binding:"required"`
	Email string `gorm:"uniqueIndex;not null" form:"email" binding:"required"`
}
