package model

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DBEngine struct {
	*gorm.DB

	userName string
	passWord string
	ip       string
	port     string
	name     string
}

func NewDBEngine() *DBEngine {
	return &DBEngine{}
}

func (db *DBEngine) init() error {
	migrator := db.Migrator()

	if !migrator.HasTable(&User{}) || !migrator.HasTable(&Password{}) || !migrator.HasTable(&Repository{}) {
		return db.AutoMigrate(&User{}, &Password{}, &Repository{})
	}
	return nil
}

func (db *DBEngine) WithUserName(userName string) *DBEngine {
	db.userName = userName
	return db
}

func (db *DBEngine) WithPassword(password string) *DBEngine {
	db.passWord = password
	return db
}

func (db *DBEngine) WithIp(ip string) *DBEngine {
	db.ip = ip
	return db
}

func (db *DBEngine) WithPort(port string) *DBEngine {
	db.port = port
	return db
}

func (db *DBEngine) WithDBName(name string) *DBEngine {
	db.name = name
	return db
}

func (db *DBEngine) Run() error {
	if db.userName == "" {
		return fmt.Errorf("empty db userName")
	}
	if db.passWord == "" {
		return fmt.Errorf("empty db password")
	}
	if db.name == "" {
		return fmt.Errorf("empty db name")
	}
	if db.port == "" {
		return fmt.Errorf("empty db port")
	}
	if db.ip == "" {
		return fmt.Errorf("empty db ip")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", db.userName, db.passWord, db.ip, db.port, db.name)
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	db.DB = database
	err = db.init()
	if err != nil {
		return err
	}

	log.Infof("database(%s) (%s:%s) connect successes!", db.name, db.ip, db.port)
	return nil
}
