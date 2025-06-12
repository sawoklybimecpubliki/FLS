package core

import (
	"database/sql"
	"errors"
	"mime/multipart"
)

var (
	ErrInsert  = errors.New("insert error")
	ErrUpdate  = errors.New("update error")
	ErrDelete  = errors.New("delete error")
	ErrSelect  = errors.New("no such file")
	ErrAddLink = errors.New("link create error")
	ErrGetLink = errors.New("link unreachable")
	ErrGetFile = errors.New("get error")
)

type LinkJSON struct {
	LinkID         string `json:"linkID,omitempty"`
	Filename       string `json:"filename,omitempty"`
	NumberOfVisits string `json:"numberOfVisits,omitempty"`
	Lifetime       string `json:"lifetime,omitempty"`
}

type Product struct {
	StorageID string
	Filename  string
	FileID    string
}

type Element struct {
	Filename string `json:"Filename"`
	Size     int64
	F        multipart.File
}

type Link struct {
	FileID         string
	LinkID         string
	NumberOfVisits sql.NullInt16
	Lifetime       sql.NullTime
}
