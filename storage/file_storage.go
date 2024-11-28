package storage

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

const (
	limitWeight = 209715200
)

type Element struct {
	Filename string         `json:"Filename"`
	Size     int64          `json:"omitempty"`
	F        multipart.File `json:"omitempty"`
}

type StoreFiles struct {
	storePath string
}

func (sf *StoreFiles) UploadFile(ctx context.Context, file Element, id string) error {

	if file.Size > limitWeight {
		return errors.New("File too heavy")
	}

	path := filepath.Join("./storage/file_storage", id)

	errC := os.Chdir(path)

	log.Println("C: ", errC)

	if errC != nil {
		errM := os.Mkdir(path, fs.ModeDir)
		if errM != nil {
			log.Println("error creating directory", errM)
			return errM
		}
	}

	dst, errD := os.Create(file.Filename)

	if errD != nil {
		log.Println("error creating file:", errD)
		return errD
	}

	defer dst.Close()
	//TODO происходить переписывание файла, либо запретить одинаковые названия, либо обходить переписывание
	_, err := io.Copy(dst, file.F)

	if err != nil {
		log.Fatal("file: ", err)
		return err
	}
	log.Println("VSE OK")
	return nil
}

func (sf *StoreFiles) DeleteFile(ctx context.Context, id string, name string) error {

	err := os.Remove(filepath.Join("./file_storage", id, name))

	if err != nil {
		log.Println("deleting failed")
		return err
	}

	return nil
}

func (sf *StoreFiles) SelectFile(ctx context.Context, id string, filename string) (Element, error) {
	var file Element
	f, err := os.Open("./file_storage" + id + filename)

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
