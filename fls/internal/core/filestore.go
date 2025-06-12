package core

import (
	"database/sql"
	"log"
)

type FileDB struct {
	Db *sql.DB
}

func (s *FileDB) SelectALL() []Product {
	rows, err := s.Db.Query("select storageid, file, fileid from Files")
	if err != nil {
		log.Println(err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Println("error row close", err)
		}
	}(rows)
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
	return product
}

func (s *FileDB) Select(idStorage, filename string) (Product, error) {
	rows := s.Db.QueryRow("select storageid, file, fileid from Files WHERE storageID=$1 AND file=$2", idStorage, filename)

	p := Product{}

	err := rows.Scan(&p.StorageID, &p.Filename, &p.FileID)
	if err != nil {
		log.Println("Select: ", err)
		return Product{}, ErrSelect
	}

	return p, nil
}

func (s *FileDB) SelectByID(fileId string) (Product, error) {
	rows := s.Db.QueryRow("select storageid, file, fileid from Files WHERE fileID=$1", fileId)

	p := Product{}

	err := rows.Scan(&p.StorageID, &p.Filename, &p.FileID)
	if err != nil {
		log.Println("Select: ", err)
		return Product{}, err
	}

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

func (s *FileDB) CountFiles(storageID string) int {
	rows, err := s.Db.Query("select count(*) from Files where storageid=$1", storageID)
	if err != nil {
		log.Println(err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Println("error row close", err)
		}
	}(rows)
	var count int
	rows.Next()
	err = rows.Scan(&count)
	if err != nil {
		log.Println("Error scan rows", err)
	}
	return count
}
