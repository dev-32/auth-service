package password

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const minCost = bcrypt.DefaultCost

func Hash(plain string) (string, error) {
	if len(plain) < 8 {
		return "", errors.New("password must be at least 8 characters")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), minCost)

	if err != nil {
		return "", nil
	}

	return string(hashed), nil
}

func Verify(plain, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
	return err == nil
}
