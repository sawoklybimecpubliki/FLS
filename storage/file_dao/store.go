package file_dao

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"time"
)

var (
	ErrInsert = errors.New("Insert error")
	ErrUpdate = errors.New("Update error")
	ErrDelete = errors.New("Delete error")
	ErrSelect = errors.New("No such file")
)

type FiLiInfo struct {
	F FileDAO
	L LinksDAO
}

type FileDAO interface {
	Select(idStorage, filename string) (Product, error)
	SelectByID(fileID string) (Product, error)
	Insert(p Product) error
	Update(idStorage, filename, newFilename string) error
	Delete(idStorage, filename string) error
}

type Product struct {
	StorageID string
	Filename  string
	FileID    string
}

type FileDB struct {
	Db *sql.DB
}

type Config struct {
	User       string
	Password   string
	Host       string
	Port       string
	DBName     string
	SSLMode    string
	DriverName string
}

func NewClient(cfg Config) (*sql.DB, error) {
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

func (s *FileDB) SelectALL() {
	rows, err := s.Db.Query("select * from Files")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	product := []Product{}
	for rows.Next() {
		p := Product{}
		err := rows.Scan(&p.StorageID, &p.Filename, &p.FileID)
		if err != nil {
			log.Println("Scan: ", err)
			continue
		}
		product = append(product, p)
	}
	for _, p := range product {
		fmt.Println(p.StorageID, p.Filename, p.FileID)
	}
}

func (s *FileDB) Select(idStorage, filename string) (Product, error) {
	rows := s.Db.QueryRow("select * from Files WHERE storageID=$1 AND file=$2", idStorage, filename)

	p := Product{}

	err := rows.Scan(&p.StorageID, &p.Filename, &p.FileID)
	if err != nil {
		log.Println("Select: ", err)
		return Product{}, ErrSelect
	}

	fmt.Println(p.StorageID, p.Filename, p.FileID)
	return p, nil
}

func (s *FileDB) SelectByID(fileId string) (Product, error) {
	rows := s.Db.QueryRow("select * from Files WHERE fileID=$1", fileId)

	p := Product{}

	err := rows.Scan(&p.StorageID, &p.Filename, &p.FileID)
	if err != nil {
		log.Println("Select: ", err)
		return Product{}, err
	}

	fmt.Println(p.StorageID, p.Filename, p.FileID)
	return p, nil
}

func (s *FileDB) Insert(p Product) error {
	_, err := s.Db.Exec("insert into Files(storageID, file, fileID) values ($1,$2,$3)",
		p.StorageID, p.Filename, p.FileID)

	if err != nil {
		log.Println("error exec: ", err)
		return ErrInsert
	}
	return nil
}

func (s *FileDB) Update(idStorage, filename, newFilename string) error {
	_, err := s.Db.Exec("update Files set file=?  WHERE file=? AND idstorage=?",
		newFilename, filename, idStorage)
	if err != nil {
		log.Println("error exec: ", err)
		return ErrUpdate
	}
	return nil
}

func (s *FileDB) Delete(idStorage, filename string) error {
	_, err := s.Db.Exec("delete from Files WHERE storageID=$1 AND file=$2", idStorage, filename)
	if err != nil {
		log.Println("error exec: ", err)
		return ErrDelete
	}
	return nil
}
