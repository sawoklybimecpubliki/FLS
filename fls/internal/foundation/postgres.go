package foundation

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"time"
)

type PGSConfig struct {
	User       string
	Password   string
	Host       string
	Port       string
	DBName     string
	SSLMode    string
	DriverName string
}

func NewPGSClient(cfg PGSConfig) (*sql.DB, error) {
	conSTR := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
	db, _ := sql.Open(cfg.DriverName, conSTR)
	flag := true
	for range 4 {
		if err := db.Ping(); err != nil {
			log.Println("error connecting to pgs: ", err)
			flag = false
		} else {
			flag = true
			break
		}
		time.Sleep(1 * time.Second)
	}
	if !flag {
		return nil, errors.New("error connecting to pgs")
	}
	log.Println("connections succeed")
	return db, nil
}
