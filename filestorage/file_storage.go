package filestorage

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
)

var (
	ErrIsExist  = errors.New("Already exist ")
	ErrTooHeavy = errors.New("File too heavy ")
)

const (
	limitSize = 209715200
	//storagePath = "C:/Users/sawok/GolandProjects/cloudService/storage/file_storage"
	//storagePath = "/app/storage/file_storage"
)

type Service interface {
	UploadFile(ctx context.Context, file Element, id string) error
	DeleteFile(ctx context.Context, id string, name string) error
	SelectFile(ctx context.Context, id string, filename string) (Element, error)
}

type Element struct {
	Filename string `json:"Filename"`
	Size     int64
	F        multipart.File
}

type StoreFiles struct {
	StorePath string
}

func (sf *StoreFiles) UploadFile(ctx context.Context, file Element, storageId string) error {

	if file.Size > limitSize {
		return ErrTooHeavy
	}

	path := filepath.Join(sf.StorePath, storageId)
	if !checkFileExists(path) {
		errM := os.Mkdir(path, fs.ModeDir)
		if errM != nil {
			log.Println("error creating directory", errM)
			return errM
		}
	}

	if checkFileExists(filepath.Join(path, file.Filename)) {
		return ErrIsExist
	}

	dst, err := os.Create(filepath.Join(path, file.Filename))

	if err != nil {
		log.Println("error creating file:", err)
		return err
	}

	defer dst.Close()

	_, err = io.Copy(dst, file.F)

	if err != nil {
		log.Fatal("file: ", err)
		return err
	}
	return nil
}

func (sf *StoreFiles) DeleteFile(ctx context.Context, storageId string, name string) error {

	path := filepath.Join(sf.StorePath, storageId, name)

	err := os.Remove(path)

	if err != nil {
		log.Println("deleting failed")
		return err
	}

	return nil
}

func (sf *StoreFiles) SelectFile(ctx context.Context, storageId string, filename string) (Element, error) {

	var file Element
	path := filepath.Join(sf.StorePath, storageId, filename)
	f, err := os.Open(path)
	defer f.Close()

	if err != nil {
		log.Println("no such file in storage")
		return Element{}, err
	}

	size, _ := f.Stat()

	file.F = f
	file.Size = size.Size()
	file.Filename = filename

	return file, nil
}

func checkFileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
