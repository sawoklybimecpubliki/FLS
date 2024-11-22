package main

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"service/api"
	"service/storage"
)

type Config struct {
	Host           string
	Port           int
	TimeoutSeconds int
	DBName         string
}

// TODO отдельная функция run

func main() {

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://mongodb:27017"))
	if err != nil {
		log.Println("Error connection to mongodb")
		return
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Println(err)
		}
	}()

	store := storage.UserDAO{client.Database("cloud").Collection("Users")}
	//store := storage.DataBase{make(map[string]string)}

	hand := api.Handler{&store}

	router := http.NewServeMux()
	hand.Mux(router)

	log.Println("server start listening on :80")
	err = http.ListenAndServe(":80", router)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
