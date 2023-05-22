package cryto

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CheckPasswordHash(password, hash string) bool {
	hashedPassword := []byte(hash)
	err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))

	log.Debugf("request:%s db-passwd:%s err:%v", password, hash, err)

	return err == nil
}
