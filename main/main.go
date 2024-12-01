package main

import (
	"FLS/api"
	"FLS/filestorage"
	"FLS/storage"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
)

type Config struct {
	Host           string
	Port           int
	TimeoutSeconds int
	DBName         string
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://mongodb:27017"))
	if err != nil {
		return err
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Println(err)
		}
	}()

	store := storage.UserDAO{client.Database("cloud").Collection("Users")}
	filestore := filestorage.StoreFiles{}

	//store := storage.DataBase{make(map[string]string)}

	hand := api.Handler{&store, &filestore}

	router := http.NewServeMux()
	hand.Mux(router)

	log.Println("server start listening on :80")

	return http.ListenAndServe(":80", router)
}
