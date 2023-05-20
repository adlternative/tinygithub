package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Name  string `gorm:"uniqueIndex;not null;size:24" form:"username"`
	Email string `gorm:"uniqueIndex;not null;size:24" form:"email"`

	Repositories []Repository
}
