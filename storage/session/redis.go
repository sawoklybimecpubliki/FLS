package session

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"log"
	"time"
)

type Config struct {
	Addr        string
	Password    string
	DB          int
	MaxRetries  int
	DialTimeout time.Duration
	Timeout     time.Duration
}

func NewClient(ctx context.Context, cfg Config) (*redis.Client, error) {
	db := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout * time.Second,
		ReadTimeout:  cfg.Timeout * time.Second,
		WriteTimeout: cfg.Timeout * time.Second,
	})

	if err := db.Ping().Err(); err != nil {
		fmt.Printf("failed to connect to redis server: %s\n", err.Error())
		return nil, err
	}

	return db, nil
}

type SessionStore struct {
	R *redis.Client
}

func (r *SessionStore) Set(p Provider) string {
	sessionId := uuid.NewString()
	jsonBytes, err := json.Marshal(p)
	if err != nil {
		log.Println("Marshaling error")
		return ""
	}

	r.R.Set(sessionId, jsonBytes, 25*time.Hour)
	return sessionId
}

func (r *SessionStore) Get(sessionId string) (Provider, error) {
	var out Provider
	result, err := r.R.Get(sessionId).Result()
	if err != nil {
		log.Println("Error get from redis")
		return Provider{}, err
	}
	err = json.Unmarshal([]byte(result), &out)
	if err != nil {
		log.Println("Unmarshalling error")
		return Provider{}, err
	}
	return out, nil
}

func (r *SessionStore) Del(sessionId string) {
	r.R.Del(sessionId)
}

func (r *SessionStore) Exists(sessionId string) (string, bool) {
	err := r.R.Get(sessionId)
	if err.Err() != redis.Nil {
		return sessionId, true
	}
	return err.String(), false
}
