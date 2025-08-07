package user

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
)

type Storer interface {
	InsertUser(ctx context.Context, user User) error
	GetUser(ctx context.Context, u User) (User, error)
	DeleteUser(ctx context.Context, u User) error
}

type SessionStorer interface {
	AddSession(ctx context.Context, ses Session) error
	CheckSession(ctx context.Context, ses Session) error
	ExtendSession(ctx context.Context, ses Session) error
	DeleteSession(ctx context.Context, ses Session) error
}

type StatStorer interface {
	GetStats(ctx context.Context) ([]Stat, error)
}

type Service struct {
	userStore    Storer
	sessionStore SessionStorer
	statStore    StatStorer
}

func NewService(userStore Storer, sessionStore SessionStorer, statStore StatStorer) *Service {
	return &Service{
		userStore:    userStore,
		sessionStore: sessionStore,
		statStore:    statStore,
	}
}

func (s *Service) findExistingUser(ctx context.Context, u User) error {
	if _, err := s.userStore.GetUser(ctx, u); err == nil {
		return errors.New("uncorrected username")
	}
	return nil
}

func (s *Service) Register(ctx context.Context, user User) error {
	log.Println("user", user)
	if err := user.Encrypt(); err != nil {
		return fmt.Errorf("error encrypting password: %w", err)
	}
	user.IdStorage = user.Login + "_storage"

	if err := s.userStore.InsertUser(ctx, user); err != nil {
		return fmt.Errorf("error inserting user: %w", err)
	}

	return nil
}

func (s *Service) Login(ctx context.Context, user User) (*Session, error) {
	existingUser, err := s.userStore.GetUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	if err := user.CheckPassword(existingUser); err != nil {
		return nil, fmt.Errorf("error checking password: %w", err)
	}

	sid, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("error generating session id: %w", err)
	}

	ses := Session{
		SID:      sid,
		Username: user.Login,
	}

	if err := s.sessionStore.AddSession(ctx, ses); err != nil {
		return nil, fmt.Errorf("error adding session: %w", err)
	}

	return &ses, nil
}

func (s *Service) CheckAuth(ctx context.Context, session Session) error {
	if err := s.sessionStore.CheckSession(ctx, session); err != nil {
		return fmt.Errorf("error checking session: %w", err)
	}

	if err := s.sessionStore.ExtendSession(ctx, session); err != nil {
		return fmt.Errorf("error extending session: %w", err)
	}

	return nil
}

func (s *Service) Logout(ctx context.Context, session Session) error {
	if err := s.sessionStore.DeleteSession(ctx, session); err != nil {
		return fmt.Errorf("error deleting session: %w", err)
	}

	return nil
}

func (s *Service) GetStats(ctx context.Context) ([]Stat, error) {
	result, err := s.statStore.GetStats(ctx)
	if err != nil {
		log.Println("error get stat", err)
		return []Stat{}, err
	}
	return result, nil
}
