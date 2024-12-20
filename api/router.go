package api

import (
	"FLS/filestorage"
	"FLS/storage"
	"FLS/storage/file_dao"
	"FLS/storage/session"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log"
	"mime/multipart"
	"net/http"
)

type Handler struct {
	App      Service
	Sessions session.Session
}

type LinkJSON struct {
	LinkID         string `json:"linkID,omitempty"`
	Filename       string `json:"filename,omitempty"`
	NumberOfVisits string `json:"numberOfVisits,omitempty"`
	Lifetime       string `json:"lifetime,omitempty"`
}

type Response struct {
	Data   any `json:"data"`
	Status any `json:"status"`
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

	if err := h.App.Users.AddNewUser(r.Context(), u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	Respond("New user was add in database", "Success", w)
}

func (h *Handler) ShowAll(w http.ResponseWriter, r *http.Request) {
	var u []storage.User

	u, _ = h.App.Users.All(r.Context())
	if u == nil {
		Respond("users not found", nil, w)
		return
	}

	Respond(u, "Success", w)
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
		return
	}
	if err := h.App.Users.Authentication(r.Context(), u); err != nil {
		Respond("Error", err, w)
		return
	}

	sessionId, err := h.Sessions.StartSession(u.Login, u.Login+"_storage")
	if err != nil {
		_, _ = fmt.Fprint(w, err)
		return
	}

	w.Header().Add("Session", sessionId)
	Respond("Successful login", "Success", w)
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
		return
	}

	if err := h.App.Users.DeleteUser(r.Context(), u.Login); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	Respond("User was delete from database", "Success", w)
}

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {

	file, header, _ := r.FormFile("file")
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			log.Println("error file close", err)
		}
	}(file)

	sessionId := r.Header.Get("Session")
	storageId := h.Sessions.GetIdStorage(sessionId)
	err := h.App.FilesStorage.UploadFile(
		r.Context(),
		filestorage.Element{
			Filename: header.Filename,
			Size:     header.Size,
			F:        file,
		},
		storageId,
	)
	if err != nil {
		log.Println(err)
		_, _ = fmt.Fprint(w, err)
		return
	}
	err = h.App.FiLiInfo.F.Insert(
		file_dao.Product{
			StorageID: storageId,
			Filename:  header.Filename,
			FileID:    uuid.NewString(),
		},
	)
	if err != nil {
		log.Println(err)
		_, _ = w.Write([]byte("Error ins"))
		return
	}

	Respond("File uploaded successfully", "Success", w)
}

func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {

	filename := r.PathValue("id")
	sessionId := r.Header.Get("Session")
	storageId := h.Sessions.GetIdStorage(sessionId)

	file, _ := h.App.FiLiInfo.F.Select(storageId, filename)
	if err := h.App.FiLiInfo.L.DeleteAllLinks(file.FileID); err != nil {
		log.Println("error deleting links", err)
		Respond(http.StatusBadRequest, errors.New("Error"), w)
		return
	}

	if err := h.App.FiLiInfo.F.Delete(storageId, filename); err != nil {
		log.Println("Error del")
		Respond(http.StatusBadRequest, errors.New("Error "), w)
		return
	}

	if err := h.App.FilesStorage.DeleteFile(r.Context(), storageId, filename); err != nil {
		log.Println("Deleting file error", err)
		Respond(http.StatusBadRequest, errors.New("Error "), w)
		return
	}

	Respond("Successful delete", "Success", w)

}

func (h *Handler) GetFile(w http.ResponseWriter, r *http.Request) {

	sessionId := r.Header.Get("Session")
	storageId := h.Sessions.GetIdStorage(sessionId)

	filename := r.PathValue("id")
	e, err := h.App.FilesStorage.SelectFile(r.Context(), storageId, filename)
	if err != nil {
		log.Println("Get failed")
		Respond(http.StatusBadRequest, errors.New("Error "), w)
		return
	}
	defer e.F.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+e.Filename)
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))

	_, err = io.Copy(w, e.F)
	if err != nil {
		log.Println("error send file to user: ", err)
		Respond(http.StatusBadRequest, errors.New("Error "), w)
	}

}

func (h *Handler) AddLinkForFile(w http.ResponseWriter, r *http.Request) {

	var l LinkJSON
	var s []byte

	s, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(s, &l)
	if err != nil {
		Respond(http.StatusBadRequest, errors.New("Error "), w)
		return
	}
	sessionId := r.Header.Get("Session")
	storageId := h.Sessions.GetIdStorage(sessionId)

	link, err := h.App.AddLink(storageId, l)
	if err != nil {
		Respond(http.StatusBadRequest, err, w)
		return
	}

	Respond("/link/"+link, "Success", w)
}

func (h *Handler) GetFileFromLink(w http.ResponseWriter, r *http.Request) {
	linkID := r.PathValue("id")

	e, err := h.App.GetFile(r.Context(), linkID)
	if err != nil {
		log.Println("Error get", err)
		Respond(http.StatusBadRequest, "No such file", w)
		return
	}
	defer e.F.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+e.Filename)
	w.Header().Set("Content-Type", "multipart/form-data")
	_, err = io.Copy(w, e.F)
	if err != nil {
		log.Println("error send file to user: ", err)
		Respond(http.StatusBadRequest, errors.New("Error send file "), w)
		return
	}
}

func (h *Handler) DeleteLink(w http.ResponseWriter, r *http.Request) {
	var l LinkJSON
	var s []byte

	s, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(s, &l)
	if err != nil {
		Respond(http.StatusBadRequest, errors.New("Error "), w)
		return
	}

	sessionId := r.Header.Get("Session")
	storageId := h.Sessions.GetIdStorage(sessionId)

	if err := h.App.DeleteLink(r.Context(), storageId, l); err != nil {
		Respond(http.StatusBadRequest, err, w)
		return
	}

	Respond("link was delete", "Success", w)
}

func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionId := r.Header.Get("Session")
		f, err := h.Sessions.CheckSession(sessionId)
		if err != nil {
			_, _ = fmt.Fprint(w, err)
			return
		}
		log.Println("refresh: ", f)
		if !f {
			_, err = h.Sessions.SessionRefresh(sessionId)
			if err != nil {
				log.Println("error refresh: ", err)
			}
			w.Header().Set("Session", sessionId)
		}
		next.ServeHTTP(w, r)
	}
}

func Respond(answer any, status any, w http.ResponseWriter) {
	out, err := json.Marshal(Response{
		Data:   answer,
		Status: status,
	})
	if err != nil {
		log.Println("marshal err: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write(out)
}

func (h *Handler) Mux(mux *http.ServeMux) {
	mux.HandleFunc("GET /users", h.AuthMiddleware(h.ShowAll))
	mux.HandleFunc("POST /user", h.Registration)
	mux.HandleFunc("GET /user", h.Login)
	mux.HandleFunc("POST /file", h.AuthMiddleware(h.UploadFile))
	mux.HandleFunc("DELETE /file/{id}", h.AuthMiddleware(h.DeleteFile))
	mux.HandleFunc("GET /file/{id}", h.AuthMiddleware(h.GetFile))
	mux.HandleFunc("POST /link", h.AuthMiddleware(h.AddLinkForFile))
	mux.HandleFunc("GET /link/{id}", h.GetFileFromLink)
	mux.HandleFunc("DELETE /link", h.AuthMiddleware(h.DeleteLink))
}
