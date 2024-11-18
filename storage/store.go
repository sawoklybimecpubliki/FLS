package storage

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
)

type Database interface {
	AddNewUser(u User) error
	Del(id string) error
	All() []User
	Authentication(u User) error
}

type User struct {
	Login    string `json:"Login"`
	Password string `json:"Password"`
}

func (u *User) hash() ([]byte, error) {
	passHash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("failed to generate password hash", err)
		return []byte{}, err
	}
	return passHash, nil
}

type DataBase struct {
	Data map[string][]byte
}

// возможно переделать или убрать вообшще, в Authentication она не используется
func (d *DataBase) compareLogin(loginNewUser string) error {
	for login := range d.Data {
		if login == loginNewUser {
			return errors.New("Uncorrect username") // отдать пользователю ошибку, но не сообщать что есть такой пользователь
		}
	}
	return nil
}

func (d *DataBase) AddNewUser(u User) error {
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

func (d *DataBase) Authentication(u User) error {
	if _, ok := d.Data[u.Login]; !ok {
		return errors.New("user not found")
	}
	if err := bcrypt.CompareHashAndPassword(d.Data[u.Login], []byte(u.Password)); err != nil {
		return errors.New("invalid password")
	}

	return nil
}

func (d *DataBase) Del(id string) error {
	if _, ok := d.Data[id]; !ok {
		return errors.New("user not found")
	}
	delete(d.Data, id)
	return nil
}

func (d *DataBase) All() []User {
	var u []User
	for login, password := range d.Data {
		u = append(u, User{login, string(password)})
	}
	return u
}
