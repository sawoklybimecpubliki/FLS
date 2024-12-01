package storage

import (
	"FLS/filestorage"
	"context"
	"errors"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
)

type Database interface {
	AddNewUser(ctx context.Context, u User) error
	DeleteUser(ctx context.Context, id string) error
	All(ctx context.Context) ([]User, error)
	Authentication(ctx context.Context, u User) (string, error)
}

type FileStorage interface {
	UploadFile(ctx context.Context, file filestorage.Element, id string) error
	DeleteFile(ctx context.Context, id string, name string) error
	SelectFile(ctx context.Context, id string, filename string) (filestorage.Element, error)
}

type User struct {
	Login     string    `json:"Login" bson:"login"`
	Password  string    `json:"Password" bson:"password"`
	IdStorage uuid.UUID `json:"omitempty" bson:"storageId"`
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
			return errors.New("Uncorrect username") // отдать пользователю ошибку, но не сообщать что есть такой пользователь
		}
	}
	return nil
}

func (d *DataBase) AddNewUser(ctx context.Context, u User) error {
	if err := d.compareLogin(u.Login); err != nil {
		return err
	}
	if p, err := u.hash(); err == nil {
		d.Data[u.Login] = p
	} else {
		return err
	}
	return nil
}

func (d *DataBase) Authentication(ctx context.Context, u User) (*jwt.Token, error) {
	if _, ok := d.Data[u.Login]; !ok {
		return nil, errors.New("user not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(d.Data[u.Login]), []byte(u.Password)); err != nil {
		return nil, errors.New("invalid password")
	}
	//TODO тут jwt token
	return nil, nil
}

func (d *DataBase) DeleteUser(ctx context.Context, id string) error {
	if _, ok := d.Data[id]; !ok {
		return errors.New("user not found")
	}
	delete(d.Data, id)
	return nil
}

func (d *DataBase) All(ctx context.Context) ([]User, error) {
	var u []User
	for login, password := range d.Data {
		u = append(u, User{login, string(password), uuid.Nil})
	}
	return u, nil
}
