package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/sawoklybimecpubliki/fistoli/filestorage"
	"github.com/sawoklybimecpubliki/fistoli/storage"
	"github.com/sawoklybimecpubliki/fistoli/storage/file_dao"
	"github.com/sawoklybimecpubliki/fistoli/storage/session"
	"io"
	"log"
	"net/http"
)

type Handler struct {
	Users        storage.Database
	FilesStorage filestorage.Service
	Sessions     session.Session
	FilesData    file_dao.FileDAO
}

func (h *Handler) Registration(w http.ResponseWriter, r *http.Request) {
	var u storage.User
	var s []byte

	s, err := io.ReadAll(r.Body)

	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(s, &u)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if err := h.Users.AddNewUser(context.Background(), u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		answer, _ := json.Marshal("New user was add in database")
		w.Write(answer)
	}
}

func (h *Handler) ShowAll(w http.ResponseWriter, r *http.Request) {
	var u []storage.User

	u, _ = h.Users.All(context.Background())
	log.Println("show all: ", u)
	if u == nil {
		answer, err := json.Marshal("users not found")
		if err != nil {
			log.Fatal(err)
		}
		w.Write(answer)
		return
	}

	answer, err := json.Marshal(u)

	if err != nil {
		log.Fatal(err)
	}
	w.Write(answer)
}

func (h *Handler) GetAnswer(w http.ResponseWriter, r *http.Request) {
	answer, err := json.Marshal("Vse rabotaet, vot otvet")
	log.Println(r.URL.Query().Get("id"))
	if err != nil {
		log.Fatal(err)
	}

	w.Write(answer)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var s []byte
	s, err := io.ReadAll(r.Body)

	if err != nil {
		log.Fatal(err)
	}

	var u storage.User
	err = json.Unmarshal(s, &u)

	if err != nil