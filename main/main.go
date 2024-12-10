package main

import (
	"FLS/api"
	"FLS/filestorage"
	"FLS/storage"
	"FLS/storage/session"
	"context"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

type AppConfig struct {
	mongo storage.Config
	redis session.Config
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var cfg AppConfig

	InitConfig()
	err := viper.Sub("mongo").Unmarshal(&cfg.mongo)
	if err != nil {
		log.Fatalf("Error reading config: %s", err)
	}

	err = viper.Sub("redis").Unmarshal(&cfg.redis)
	if err != nil {
		log.Fatalf("Error reading config: %s", err)
	}

	client, err := storage.NewClient(cfg.mongo)
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Println(err)
		}
	}()

	store := storage.UserDAO{client.Database(cfg.mongo.DBName).Collection("Users")}
	//store := storage.DataBase{make(map[string]string)}

	filestore := filestorage.StoreFiles{}

	var redis session.Redis
	redis.R, err = session.NewClient(context.Background(), cfg.redis)

	//sessionStore := session.Service{&session.Store{make(map[string]session.Provider)}}
	sessionStore := session.Service{&redis}

	hand := api.Handler{&store, &filestore, sessionStore}

	router := http.NewServeMux()
	hand.Mux(router)

	log.Println("server start listening on :80")

	return http.ListenAndServe(":80", router)
}

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./storage/")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config: %s", err)
	}
}
