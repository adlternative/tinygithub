package model

import "gorm.io/gorm"

type Repository struct {
	gorm.Model
	UserID uint
	Name   string
	Desc   string
}
