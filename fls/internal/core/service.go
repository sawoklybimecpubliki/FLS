package core

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"log"
	"strconv"
	"time"
)

type FileStorer interface {
	Select(idStorage, filename string) (Product, error)
	SelectByID(fileID string) (Product, error)
	Insert(p Product) error
	Update(idStorage, filename, newFilename string) error
	Delete(idStorage, filename string) error
	CountFiles(storageID string) int
}

type LinkStorer interface {
	Select(linkID string) (Link, error)
	SelectAllLinks(fileID string) ([]Link, error)
	Insert(l Link) error
	Update(l Link) error
	Delete(linkID string) error
	DeleteAllLinks(fileID string) error
	DeleteInvalidLinks() error
}

type Service struct {
	FileStore FileStorer
	LinkStore LinkStorer
	Storage   FileStorage
}

func NewService(fileStore FileStorer, linkStore LinkStorer, store FileStorage) *Service {
	return &Service{
		FileStore: fileStore,
		LinkStore: linkStore,
		Storage:   store,
	}
}

func (s *Service) AddLink(storageId string, link LinkJSON) (string, error) {

	file, err := s.FileStore.Select(storageId, link.Filename)
	if err != nil {
		log.Println("no such file: ", err)
		return "", ErrAddLink
	}
	linkID := uuid.NewString()
	var sqlVisits sql.NullInt16
	var sqlLifetime sql.NullTime

	if link.NumberOfVisits != "" {
		tmp, _ := strconv.Atoi(link.NumberOfVisits)
		sqlVisits.Int16 = int16(tmp)
		sqlVisits.Valid = true
	} else {
		sqlVisits.Valid = false
	}

	if link.Lifetime != "" {
		tmp, _ := strconv.Atoi(link.Lifetime)
		sqlLifetime.Time = time.Unix(time.Now().Unix()+int64(tmp), 0)
		sqlLifetime.Valid = true
	} else {
		sqlLifetime.Valid = false
	}

	if err := s.LinkStore.Insert(
		Link{
			FileID:         file.FileID,
			LinkID:         linkID,
			NumberOfVisits: sqlVisits,
			Lifetime:       sqlLifetime,
		},
	); err != nil {
		log.Println("error insert link: ", err)
		return "", ErrAddLink
	}
	return linkID, nil
}

func (s *Service) DeleteLink(ctx context.Context, storageID string, link LinkJSON) error {
	linkInfo, _ := s.LinkStore.Select(link.LinkID)
	file, err := s.FileStore.SelectByID(linkInfo.FileID)
	if errors.Is(err, ErrSelect) || file.StorageID != storageID {
		log.Println(err)
		return ErrDelete
	}

	if err := s.LinkStore.Delete(link.LinkID); err != nil {
		log.Println("error deleting link", err)
		return ErrDelete
	}
	return nil
}

func (s *Service) GetFile(ctx context.Context, linkID string) (Element, error) {

	var e Element
	linkInfo, err := s.LinkStore.Select(linkID)

	if err != nil {
		log.Println("Error select file", err)
		return Element{}, ErrSelect
	}
	if (linkInfo.NumberOfVisits.Int16 > 0 || !linkInfo.NumberOfVisits.Valid) &&
		(!linkInfo.Lifetime.Valid || linkInfo.Lifetime.Time.Unix()-time.Now().Unix() > 0) {
		file, err := s.FileStore.SelectByID(linkInfo.FileID)
		if err != nil {
			log.Println("error select file:", err)
			return Element{}, ErrSelect
		}
		e, err = s.Storage.SelectFile(ctx, file.StorageID, file.Filename)
		if err != nil {
			log.Println("Get failed")
			return Element{}, ErrGetFile
		}

	} else {
		log.Println("Error: ", err)
		return Element{}, ErrGetLink
	}

	if linkInfo.NumberOfVisits.Valid && linkInfo.NumberOfVisits.Int16 > 0 {
		linkInfo.NumberOfVisits.Int16--
		err := s.LinkStore.Update(linkInfo)
		if err != nil {
			log.Println("error update link:", err)
			return Element{}, ErrUpdate
		}
	}
	return e, nil
}

func (s *Service) UploadFile(ctx context.Context, file Element, storageID string) error {

	if s.FileStore.CountFiles(storageID) >= 5 {
		return errors.New("files limit")
	}

	err := s.Storage.UploadFile(
		ctx,
		file,
		storageID,
	)
	if err != nil {
		log.Println(err)
		return err
	}
	err = s.FileStore.Insert(
		Product{
			StorageID: storageID,
			Filename:  file.Filename,
			FileID:    uuid.NewString(),
		},
	)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (s *Service) DeleteFile(ctx context.Context, storageId, filename string) error {
	file, _ := s.FileStore.Select(storageId, filename)
	if err := s.LinkStore.DeleteAllLinks(file.FileID); err != nil {
		log.Println("error deleting links", err)
		return err
	}

	if err := s.FileStore.Delete(storageId, filename); err != nil {
		log.Println("Error del")
		return err
	}

	if err := s.Storage.DeleteFile(ctx, storageId, filename); err != nil {
		log.Println("Deleting file error", err)
		return err
	}
	return nil
}
