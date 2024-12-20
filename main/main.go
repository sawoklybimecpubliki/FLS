package main

import (
	"FLS/api"
	"FLS/filestorage"
	"FLS/storage"
	"FLS/storage/file_dao"
	"FLS/storage/session"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"strings"
)

type AppConfig struct {
	Mongo    storage.Config
	Redis    session.Config
	Postgres file_dao.Config
	Migrate  Migrate_cfg
}

type Migrate_cfg struct {
	Source string
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

	filestore := filestorage.StoreFiles{StorePath: cfg.Mongo.Path}

	var redis session.SessionStore
	redis.R, err = session.NewClient(context.Background(), cfg.Redis)

	//sessionStore := session.Service{&session.Store{make(map[string]session.Provider)}}
	sessionStore := session.Service{Val: &redis}

	var filesDAO file_dao.FileDB
	var linksDAO file_dao.LinkDB

	filesDAO.Db, err = file_dao.NewClient(cfg.Postgres)
	defer func(Db *sql.DB) {
		err := Db.Close()
		if err != nil {
			fmt.Println("error close fileDB", err)
		}
	}(filesDAO.Db)
	linksDAO.Db, err = file_dao.NewClient(cfg.Postgres)
	defer func(Db *sql.DB) {
		err := Db.Close()
		if err != nil {
			log.Println("error close linkDB", err)
		}
	}(linksDAO.Db)

	hand := api.Handler{
		App: api.Service{
			Users:        &store,
			FilesStorage: &filestore,
			FiLiInfo:     file_dao.FiLiInfo{F: &filesDAO, L: &linksDAO},
		},
		Sessions: &sessionStore,
	}

	migrateDB(filesDAO.Db, cfg.Postgres.DBName, cfg.Migrate)

	router := http.NewServeMux()
	hand.Mux(router)

	log.Println("server start listening on :80")
	return http.ListenAndServe(":80", router)
}

func migrateDB(db *sql.DB, dbName string, cfg Migrate_cfg) {
	f := file.File{}
	sourceDriver, err := f.Open(cfg.Source)
	if err != nil {
		log.Println("err in source: ", err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Println("err in driver:", err)
	}
	m, err := migrate.NewWithInstance(
		"migrations",
		sourceDriver,
		dbName,
		driver)
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Println("migrate error: ", err)
		return
	}
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
