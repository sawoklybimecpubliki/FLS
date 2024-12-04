package session

import (
	"errors"
	"github.com/google/uuid"
	"log"
	"time"
)

var (
	DestroyErr = errors.New("Error of destruction ")
	SessionErr = errors.New("Session not found ")
)

type Session interface {
	StartSession(login, idStorage string) (string, error)
	DestroySession(sessionId string) error
	CheckSession(sessionId string) (bool, error)
	SessionRefresh(sessionId string) (string, error)
	GetIdStorage(sessionId string) string
}

type Provider struct {
	Login     string
	IdStorage string
	Lifetime  int64
}

type SessionStorer interface {
	Set(p Provider) string
	Get(sessionId string) (Provider, error)
	Del(sessionId string)
	Exists(login string) (string, bool)
}

type Store struct {
	Val map[string]Provider
}

func (s *Store) Set(p Provider) string {
	sessionId := uuid.NewString()
	s.Val[sessionId] = p
	return sessionId
}

func (s *Store) Get(sessionId string) (Provider, error) {
	if _, ok := s.Val[sessionId]; !ok {
		log.Println("Session doesnt exist")
		return Provider{}, SessionErr
	}
	return s.Val[sessionId], nil
}

func (s *Store) Del(sessionId string) {
	delete(s.Val, sessionId)
}

func (s *Store) IsExist(login string) (string, bool) {
	for id, value := range s.Val {
		if value.Login == login {
			log