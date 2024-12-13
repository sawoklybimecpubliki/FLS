package main

import (
	"FLS/api"
	"FLS/filestorage"
	"FLS/storage"
	"FLS/storage/file_dao"
	"FLS/storage/session"
	"context"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"strings"
)

type AppConfig struct {
	Mongo    storage.Config
	Redis    session.Config
	Postgres file_dao.Config
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var cfg AppConfig
	InitConfig()
	err := viper.Unmarshal(&cfg)
	//log.Fatal(cfg)
	if err != nil {
		log.Fatalf("Error reading config: %s", err)
	}

	client, err := storage.NewClient(cfg.Mongo)
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Println(err)
		}
	}()

	store := storage.UserDAO{
		C: client.Database(cfg.Mongo.DBName).Collection(cfg.Mongo.CollectionUsers)}
	//store := storage.DataBase{make(map[string]string)}

	filestore := filestorage.StoreFiles{cfg.Mongo.Path}

	var redis session.SessionStore
	redis.R, err = session.NewClient(context.Background(), cfg.Redis)

	//sessionStore := session.Service{&session.Store{make(map[string]session.Provider)}}
	sessionStore := session.Service{&redis}

	var filesDAO file_dao.FileDB
	filesDAO.Db, err = file_dao.NewClient(cfg.Postgres)
	defer filesDAO.Db.Close()

	hand := api.Handler{&store, &filestore, &sessionStore, &filesDAO}

	router := http.NewServeMux()
	hand.Mux(router)

	log.Println("server start listening on :80")
	//TODO добавить в конфиг хоста
	return http.ListenAndServe(":80", router)
}

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./main/")
	viper.SetEnvPrefix("FISTOLI")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config: %s", err)
	}
}
