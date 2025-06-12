package main

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"strings"
	"time"
	"us/cmd/api/router"
	"us/internal/core/user"
	"us/internal/foundation"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	log.Printf("config: %+v", cfg)

	// Mongo Client
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.App.ConnectionTimeoutSec)*time.Second)
	defer cancel()

	mongoClient, err := foundation.NewMongoClient(ctx, foundation.MongoConfig{
		Host:     cfg.Mongo.Host,
		Port:     cfg.Mongo.Port,
		Username: cfg.Mongo.Username,
		Password: cfg.Mongo.Password,
	})
	if err != nil {
		return err
	}

	// Redis Client
	redisClient, err := foundation.NewRedisClient(ctx, foundation.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
	})
	if err != nil {
		return err
	}

	sessionStore := user.NewSessionStore(
		time.Duration(cfg.Redis.SessionExpirationHours)*time.Hour,
		redisClient,
	)

	userStore, err := user.NewStore(mongoClient.Database(cfg.Mongo.DBName).Collection(cfg.Mongo.UsersCollection))
	if err != nil {
		return err
	}

	userService := user.NewService(userStore, sessionStore)

	handler := router.NewHandler(userService)

	mux := router.APIMux(handler)

	log.Println("Starting listening on port 8080")
	return http.ListenAndServe(":8080", mux)
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./cmd/api")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("US")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("could not read config: %w", err)
	}

	cfg := &Config{}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("could not unmarshal config: %w", err)
	}

	return cfg, nil
}

type Config struct {
	App struct {
		ConnectionTimeoutSec int
	}

	Mongo struct {
		Host            string
		Port            string
		Username        string
		Password        string
		DBName          string
		UsersCollection string
	}

	Redis struct {
		Host                   string
		Port                   string
		Password               string
		SessionExpirationHours int
	}
}
