package api

import (
	"FLS/filestorage"
	"FLS/storage"
	"FLS/storage/file_dao"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"log"
	"strconv"
	"time"
)

var (
	AddLinkError = errors.New("Link create error ")
	GetLinkError = errors.New("Link unreachable ")
	GetFileError = errors.New("Get error ")
)

type Service struct {
	Users        storage.Database
	FilesStorage filestorage.Service
	FiLiInfo     file_dao.FiLiInfo
}

func (s *Service) AddLink(storageId string, link LinkJSON) (string, error) {

	file, err := s.FiLiInfo.F.Select(storageId, link.Filename)
	if err != nil {
		log.Println("no such file: ", err)
		return "", AddLinkError
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

	if err := s.FiLiInfo.L.Insert(
		file_dao.Link{
			FileID:         file.FileID,
			LinkID:         linkID,
			NumberOfVisits: sqlVisits,
			Lifetime:       sqlLifetime,
		},
	); err != nil {
		log.Println("error insert link: ", err)
		return "", AddLinkError
	}
	return linkID, nil
}

func (s *Service) GetFile(ctx context.Context, linkID string) (filestorage.Element, error) {

	var e filestorage.Element
	linkInfo, err := s.FiLiInfo.L.Select(linkID)

	if err != nil {
		log.Println("Error select file", err)
		return filestorage.Element{}, file_dao.ErrSelect
	}
	if (linkInfo.NumberOfVisits.Int16 > 0 || !linkInfo.NumberOfVisits.Valid) &&
		(!linkInfo.Lifetime.Valid || linkInfo.Lifetime.Time.Unix()-time.Now().Unix() > 0) {
		file, err := s.FiLiInfo.F.SelectByID(linkInfo.FileID)
		if err != nil {
			log.Println("error select file:", err)
			return filestorage.Element{}, file_dao.ErrSelect
		}
		e, err = s.FilesStorage.SelectFile(ctx, file.StorageID, file.Filename)
		if err != nil {
			log.Println("Get failed")
			return filestorage.Element{}, GetFileError
		}

	} else {
		log.Println("Error: ", err)
		return filestorage.Element{}, GetLinkError
	}

	if linkInfo.NumberOfVisits.Valid && linkInfo.NumberOfVisits.Int16 > 0 {
		linkInfo.NumberOfVisits.Int16--
		err := s.FiLiInfo.L.Update(linkInfo)
		if err != nil {
			log.Println("error update link:", err)
			return filestorage.Element{}, file_dao.ErrUpdate
		}
	}
	return e, nil
}

func (s *Service) DeleteLink(ctx context.Context, storageID string, link LinkJSON) error {
	linkInfo, _ := s.FiLiInfo.L.Select(link.LinkID)
	file, err := s.FiLiInfo.F.SelectByID(linkInfo.FileID)
	if errors.Is(err, file_dao.ErrSelect) || file.StorageID != storageID {
		log.Println(err)
		return file_dao.ErrDelete
	}

	if err := s.FiLiInfo.L.Delete(link.LinkID); err != nil {
		log.Println("error deleting link", err)
		return file_dao.ErrDelete
	}
	return nil
}
