package user

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
)

type User struct {
	Login     string `bson:"login"`
	Password  string `bson:"password"`
	IdStorage string `bson:"storageId"`
}

func (u *User) Encrypt() error {
	passHash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("failed to generate password hash", err)
		return err
	}
	u.Password = string(passHash)
	return nil
}

func (u *User) CheckPassword(existingUser User) error {
	log.Println("check pass", existingUser.Password, u.Password)
	return bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(u.Password))
}

type Session struct {
	SID      uuid.UUID
	Username string
}
