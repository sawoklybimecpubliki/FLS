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

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	if token, err := h.Users.Authentication(context.TODO(), u); err != nil {
		answer, _ := json.Marshal(err.Error())
		w.Write(answer)
	} else {
		answer, _ := json.Marshal("Successful login")

		sessionId, err := h.Sessions.StartSession(u.Login, u.Login+"_storage")
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		w.Header().Add("Session", sessionId)

		log.Println("TOKEN: ", token)
		cookie := &http.Cookie{Name: "JWT", Value: token}
		http.SetCookie(w, cookie)
		w.Write(answer)
	}
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	var u storage.User
	s, err := io.ReadAll(r.Body)

	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(s, &u)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if err := h.Users.DeleteUser(context.Background(), u.Login); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		answer, _ := json.Marshal("User was delete from database")
		w.Write(answer)
	}

}

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {

	file, header, _ := r.FormFile("file")
	defer file.Close()
	sessionId := r.Header.Get("Session")
	storageId := h.Sessions.GetIdStorage(sessionId)
	err := h.FilesStorage.UploadFile(context.Background(),
		filestorage.Element{header.Filename, header.Size, file}, storageId)
	if err != nil {
		log.Println(err)
		fmt.Fprint(w, err)
	} else {
		err := h.FilesData.Insert(file_dao.Product{storageId, header.Filename, uuid.NewString()})
		if err != nil {
			log.Println(err)
			w.Write([]byte("Error ins"))
			return
		}
		answer, err := json.Marshal("File uploaded successfully")
		if err != nil {
			log.Println("marshaling error", err)
		}

		w.Write(answer)
	}
}

func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {

	filename := r.PathValue("id")

	sessionId := r.Header.Get("Session")
	storageId := h.Sessions.GetIdStorage(sessionId)

	if err := h.FilesStorage.DeleteFile(context.Background(), storageId, filename); err != nil {
		log.Println("Deleting file error", err)
	} else {
		err = h.FilesData.Delete(storageId, filename)
		if err != nil {
			log.Println(err)
			w.Write([]byte("Error del"))
			return
		}
		w.Write([]byte("Successful delete"))
	}
}

func (h *Handler) GetFile(w http.ResponseWriter, r *http.Request) {

	sessionId := r.Header.Get("Session")
	storageId := h.Sessions.GetIdStorage(sessionId)

	filename := r.PathValue("id")
	e, err := h.FilesStorage.SelectFile(context.Background(), storageId, filename)
	if err != nil {
		fmt.Fprint(w, "Get failed")
	} else {
		p, err := h.FilesData.Select(storageId, filename)
		if err != nil {
			log.Println(err)
			w.Write([]byte("Error sel"))
			return
		}
		w.Write([]byte("Successful select"))
		log.Println("SIZE: ", e.Size, "NAME:", e.Filename, "product: ", p)
	}
}

func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionId := r.Header.Get("Session")
		f, err := h.Sessions.CheckSession(sessionId)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		if !f {
			h.Sessions.SessionRefresh(sessionId)
		}
		next.ServeHTTP(w, r)
	}
}

func (h *Handler) Mux(mux *http.ServeMux) {
	mux.HandleFunc("GET /users", h.AuthMiddleware(h.ShowAll))
	mux.HandleFunc("GET /answer", h.GetAnswer)
	mux.HandleFunc("POST /user", h.Registration)
	mux.HandleFunc("GET /user", h.Login)
	mux.HandleFunc("POST /file", h.AuthMiddleware(h.UploadFile))
	mux.HandleFunc("DELETE /file/{id}", h.AuthMiddleware(h.DeleteFile))
	mux.HandleFunc("GET /file/{id}", h.AuthMiddleware(h.GetFile))
}
