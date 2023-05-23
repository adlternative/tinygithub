package model

import "gorm.io/gorm"

type Repository struct {
	gorm.Model
	UserID uint
	Name   string `gorm:"not null;size:24"`
	Desc   string `gorm:"not null;size:256"`
}
