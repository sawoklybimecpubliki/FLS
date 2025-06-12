package user

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type SessionStore struct {
	sessionTTL  time.Duration
	redisClient *redis.Client
}

func NewSessionStore(sessionDuration time.Duration, redisClient *redis.Client) *SessionStore {
	return &SessionStore{
		sessionTTL:  sessionDuration,
		redisClient: redisClient,
	}
}

func (s *SessionStore) AddSession(ctx context.Context, ses Session) error {
	return s.redisClient.Set(ctx, ses.SID.String(), ses.Username, s.sessionTTL).Err()
}

func (s *SessionStore) CheckSession(ctx context.Context, ses Session) error {
	value, err := s.redisClient.Get(ctx, ses.SID.String()).Result()
	if err != nil {
		return fmt.Errorf("could not get session: %w", err)
	}

	if value != ses.Username {
		return fmt.Errorf("session username does not match")
	}

	return nil
}

func (s *SessionStore) ExtendSession(ctx context.Context, ses Session) error {
	return s.redisClient.Expire(ctx, ses.SID.String(), s.sessionTTL).Err()
}

func (s *SessionStore) DeleteSession(ctx context.Context, ses Session) error {
	return s.redisClient.Del(ctx, ses.SID.String()).Err()
}
