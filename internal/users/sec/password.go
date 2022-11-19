package sec

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword - hashes the given password using bcrypt
func HashPassword(pwd string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pwd), 14)
	if err != nil {
		logrus.Errorf("Error in users/sec.HashPassword -> error: %s", err)
		return "", err
	}

	return string(hashed), err
}
