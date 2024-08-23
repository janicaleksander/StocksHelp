package user

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
)

type User struct {
	ID       uuid.UUID
	Name     string
	Email    string
	Password string
}

func NewUser(name, email, password string) *User {
	p, err := hpass(password)
	if err != nil {
		log.Print("Problem with hashing, too long passwrod probably")
		return nil
	}
	return &User{
		ID:       uuid.New(),
		Name:     name,
		Email:    email,
		Password: p,
	}
}

func hpass(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err

}
