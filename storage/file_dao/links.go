package file_dao

import (
	"database/sql"
	"fmt"
	"log"
)

type Link struct {
	FileID         string
	LinkID         string
	NumberOfVisits sql.NullInt16
	Lifetime       sql.NullTime
}

type LinksDAO interface {
	Select(linkID string) (Link, error)
	SelectAllLinks(fileID string) ([]Link, error)
	Insert(l Link) error
	Update(l Link) error
	Delete(linkID string) error
	DeleteAllLinks(fileID string) error
}

type LinkDB struct {
	Db *sql.DB
}

func (dao *LinkDB) Select(linkID string) (Link, error) {
	rows := dao.Db.QueryRow("select * from Links where linkID=$1", linkID)

	l := Link{}

	err := rows.Scan(&l.FileID, &l.LinkID, &l.NumberOfVisits, &l.Lifetime)
	if err != nil {
		log.Println("Select: ", err)
		return Link{}, err
	}

	fmt.Println(l)
	return l, nil
}

func (dao *LinkDB) SelectAllLinks(fileID string) ([]Link, error) {
	rows, err := dao.Db.Query("select * from Links where fileID=$1", fileID)
	if err != nil {
		log.Println(err)
	}
	out := []Link{}

	for rows.Next() {
		l := Link{}
		err := rows.Scan(&l.FileID, &l.LinkID, &l.NumberOfVisits, &l.Lifetime)
		if err != nil {
			log.Println("Scan: ", err)
			continue
		}
		out = append(out, l)
	}

	if err != nil {
		log.Println("Select: ", err)
		return []Link{}, err
	}

	return out, nil
}

func (dao *LinkDB) Insert(l Link) error {
	_, err := dao.Db.Exec("insert into Links(FileID, LinkID, NumberOfVisits, Lifetime) values ($1,$2,$3,$4)",
		l.FileID, l.LinkID, l.NumberOfVisits, l.Lifetime)
	if err != nil {
		log.Println("error exec: ", err)
		return ErrInsert
	}
	return nil
}

func (dao *LinkDB) Update(l Link) error {
	_, err := dao.Db.Exec("update Links set numberOfVisits=$1 where linkID=$2",
		l.NumberOfVisits, l.LinkID)
	if err != nil {
		log.Println("error exec: ", err)
		return ErrUpdate
	}
	return nil
}

func (dao *LinkDB) Delete(linkID string) error {
	_, err := dao.Db.Exec("delete from Links where LinkID=$1", linkID)
	if err != nil {
		log.Println("error exec: ", err)
		return ErrDelete
	}
	return nil
}

func (dao *LinkDB) DeleteAllLinks(fileID string) error {
	_, err := dao.Db.Exec("delete from Links where fileID=$1", fileID)
	if err != nil {
		log.Println("error exec: ", err)
		return ErrDelete
	}
	return nil
}
