package storage

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
)

type Database interface {
	AddNewUser(ctx context.Context, u User) error
	DeleteUser(ctx context.Context, id string) error
	All(ctx context.Context) ([]User, error)
	Authentication(ctx context.Context, u User) (string, error)
}

type User struct {
	Login     string `json:"Login" bson:"login"`
	Password  string `json:"Password" bson:"password"`
	IdStorage string `json:"omitempty" bson:"storageId"`
}

// TODO auth  отнести к юзеру а в базе возвращать данные пользователя

func (u *User) hash() (string, error) {
	passHash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("failed to generate password hash", err)
		return "", err
	}
	return string(passHash), nil
}

type DataBase struct {
	Data map[string]string
}

// TODO возможно переделать или убрать вообшще, в Authentication она не используется
func (d *DataBase) compareLogin(loginNewUser string) error {
	for login := range d.Data {
		if login == loginNewUser {
			return errors.New("Uncorrect username") // отдать пользователю о