package main

import (
	"encoding/json"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"log/slog"
	"net/http"
)

type Database interface {
	add(u user)
	del(id string) error
	sel() dataBase
	all() []user
}

type user struct {
	Login    string `json:"Login"`
	Password string `json:"Password"`
}

func (u *user) hash() ([]byte, error) {
	passHash, err := bcrypt.GenerateFromPasswor