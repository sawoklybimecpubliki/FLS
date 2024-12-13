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
			log.Println("session already exist")
			return id, true
		}
	}
	return "", false
}

type Service struct {
	Val SessionStorer
}

func (s *Service) StartSession(login, idStorage string) (string, error) {
	if id, f := s.Val.Exists(login); f {
		return id, nil
	}
	sessionId := s.Val.Set(Provider{login, idStorage, time.Now().Unix() + 120})
	return sessionId, nil
}

func (s *Service) DestroySession(sessionId string) error {
	if _, err := s.Val.Get(sessionId); err != nil {
		log.Println("Session not found")
		return DestroyErr
	}
	s.Val.Del(sessionId)
	return nil
}

func (s *Service) CheckSession(sessionId string) (bool, error) {
	if _, ok := s.Val.Exists(sessionId); !ok {
		log.Println(SessionErr)
		return false, SessionErr
	}
	if v, _ := s.Val.Get(sessionId); v.Lifetime <= time.Now().Unix() {
		log.Println("Lifetime: ", v.Lifetime, "Now: ", time.Now().Unix())
		log.Println("Session time is expired")
		return false, nil
	}
	return true, nil
}

func (s *Service) SessionRefresh(sessionId string) (string, error) {

	val, _ := s.Val.Get(sessionId)
	val.Lifetime = time.Now().Unix() + 120
	s.Val.Set(val)
	if err := s.DestroySession(sessionId); err != nil {
		log.Println(err)
		return "", err
	}
	return sessionId, nil
}

func (s *Service) GetIdStorage(sessionId string) string {
	p, err := s.Val.Get(sessionId)
	if err != nil {
		log.Println("Session not found")
		return ""
	}
	return p.IdStorage
}
