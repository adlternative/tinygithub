package model

import "gorm.io/gorm"

type Repository struct {
	gorm.Model
	UserID uint
	Name   string `gorm:"uniqueIndex;not null"`
	Desc   string
}
