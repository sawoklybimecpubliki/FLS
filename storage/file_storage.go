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
	storagePath = "C:/Users/sawok/GolandProjects/cloudService/storage/file_storage"
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

	os.Chdir(storagePath)

	if file.Size > limitWeight {
		return errors.New("File too heavy")
	}

	errC := os.Chdir(filepath.Join(storagePath, id))

	log.Println("C: ", errC)

	if errC != nil {
		errM := os.Mkdir(filepath.Join(storagePath, id), fs.ModeDir)
		if errM != nil {
			log.Println("error creating directory", errM)
			return errM
		}
	}
	os.Chdir(filepath.Join(storagePath, id))
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

	os.Chdir(filepath.Join("C:/Users/sawok/GolandProjects/cloudService/storage/file_storage", id))

	err := os.Remove(name)

	if err != nil {
		log.Println("deleting failed")
		return err
	}

	return nil
}

func (sf *StoreFiles) SelectFile(ctx context.Context, id string, filename string) (Element, error) {
	os.Chdir(filepath.Join("C:/Users/sawok/GolandProjects/cloudService/storage/file_storage", id))

	var file Element

	f, err := os.Open(filename)
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
