package cryto

import (
	log "github.com/sirupsen/logrus"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPasswordAndCheckPasswordHash(t *testing.T) {
	testPassword := "testpassword"

	hashedPassword, err := HashPassword(testPassword)
	assert.NoError(t, err)

	log.Printf("testPassword: %s hashedPassword: %s", testPassword, hashedPassword)

	result := CheckPasswordHash(testPassword, hashedPassword)
	assert.True(t, result)

	result = CheckPasswordHash("wrongpassword", hashedPassword)
	assert.False(t, result)
}
