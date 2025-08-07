package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sawoklybimecpubliki/FLS/fls/cmd/api/router"
	"github.com/sawoklybimecpubliki/FLS/fls/internal/core"
	"github.com/sawoklybimecpubliki/FLS/fls/internal/foundation"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

func main() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	ch := make(chan bool)
	if err := run(ch, &wg); err != nil {
		log.Fatal(err)
	}
	ch <- true
	close(ch)
	wg.Wait()
}

func run(doneCh chan bool, wg *sync.WaitGroup) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	// PGS Client
	_, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.App.ConnectionTimeoutSec)*time.Second)
	defer cancel()

	var filesDAO core.FileDB
	var linksDAO core.LinkDB
	filesDAO.Db, err = foundation.NewPGSClient(cfg.Postgres)
	if err != nil {
		log.Println("Error connect pgs: ", err)
		return err
	}
	defer func(Db *sql.DB) {
		err := Db.Close()
		if err != nil {
			log.Println("error close fileDB", err)
		}
	}(filesDAO.Db)
	linksDAO.Db, err = foundation.NewPGSClient(cfg.Postgres)
	if err != nil {
		log.Println("Error connect pgs: ", err)
		return err
	}
	defer func(Db *sql.DB) {
		err := Db.Close()
		if err != nil {
			log.Println("error close linkDB", err)
		}
	}(linksDAO.Db)

	//s3
	store := core.NewFileStore("")
	flsService := core.NewService(&filesDAO, &linksDAO, store)

	handler := router.NewHandler(flsService)

	mux := router.APIMux(handler)
	migrateDB(filesDAO.Db, cfg.Postgres.DBName, *cfg)

	go func() {
		tick := time.NewTicker(1 * time.Minute)
		for {
			select {
			case <-doneCh:
				wg.Done()
				return
			case <-tick.C:
				errD := linksDAO.DeleteInvalidLinks()
				if errD != nil {
					log.Println("Error deleting invalid links", err)
				}
			}
		}
	}()

	log.Println("Starting listening on port 8080")
	return http.ListenAndServe(":8080", mux)
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./cmd/api")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("FLS")
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

func migrateDB(db *sql.DB, dbName string, cfg Config) {
	f := file.File{}
	sourceDriver, err := f.Open(cfg.Migrate.Source)
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
	if err != nil {
		log.Println("Error migrate", err)
		return
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Println("migrate error: ", err)
		return
	}
}

type Config struct {
	App struct {
		ConnectionTimeoutSec int
	}

	Postgres struct {
		User       string
		Password   string
		Host       string
		Port       string
		DBName     string
		SSLMode    string
		DriverName string
	}
	Migrate struct {
		Source string
	}
}
