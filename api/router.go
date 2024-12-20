package api

import (
	"FLS/filestorage"
	"FLS/storage"
	"FLS/storage/file_dao"
	"FLS/storage/session"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

type Handler struct {
	Users        storage.Database
	FilesStorage filestorage.Service
	Sessions     session.Session
	FiLiInfo     file_dao.FiLiInfo
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
		answer, err := json.Marshal("New user was add in database")
		if err != nil {
			log.Println("marshal err: ", err)
		}
		_, _ = w.Write(answer)
	}
}

func (h *Handler) ShowAll(w http.ResponseWriter, r *http.Request) {
	var u []storage.User
	fmt.Println(r.Header.Get("Session"))
	u, _ = h.Users.All(context.Background())
	log.Println("show all: ", u)
	if u == nil {
		answer, err := json.Marshal("users not found")
		if err != nil {
			log.Fatal(err)
		}
		_, _ = w.Write(answer)
		return
	}

	answer, err := json.Marshal(u)

	if err != nil {
		log.Fatal(err)
	}
	_, _ = w.Write(answer)
}

func (h *Handler) GetAnswer(w http.ResponseWriter, r *http.Request) {
	answer, err := json.Marshal("Vse rabotaet, vot otvet")
	log.Println(r.URL.Query().Get("id"))
	if err != nil {
		log.Fatal(err)
	}

	_, _ = w.Write(answer)
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
		_, _ = w.Write(answer)
	} else {
		answer, _ := json.Marshal("Successful login")

		sessionId, err := h.Sessions.StartSession(u.Login, u.Login+"_storage")
		if err != nil {
			_, _ = fmt.Fprint(w, err)
			return
		}
		w.Header().Add("Session", sessionId)

		log.Println("TOKEN: ", token)
		cookie := &http.Cookie{Name: "JWT", Value: token}
		http.SetCookie(w, cookie)
		_, _ = w.Write(answer)
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
		_, _ = w.Write(answer)
	}

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
	err := h.FilesStorage.UploadFile(context.Background(),
		filestorage.Element{Filename: header.Filename, Size: header.Size, F: file}, storageId)
	if err != nil {
		log.Println(err)
		_, _ = fmt.Fprint(w, err)
	} else {
		err := h.FiLiInfo.F.Insert(file_dao.Product{StorageID: storageId, Filename: header.Filename,
			FileID: uuid.NewString()})
		if err != nil {
			log.Println(err)
			_, _ = w.Write([]byte("Error ins"))
			return
		}
		answer, err := json.Marshal("File uploaded successfully")
		if err != nil {
			log.Println("marshaling error", err)
		}

		_, _ = w.Write(answer)
	}
}

func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {

	filename := r.PathValue("id")

	sessionId := r.Header.Get("Session")
	storageId := h.Sessions.GetIdStorage(sessionId)

	if err := h.FilesStorage.DeleteFile(context.Background(), storageId, filename); err != nil {
		log.Println("Deleting file error", err)
	} else {
		file, _ := h.FiLiInfo.F.Select(storageId, filename)
		if err := h.FiLiInfo.L.DeleteAllLinks(file.FileID); err != nil {
			log.Println("error deleting links", err)
		}

		if err := h.FiLiInfo.F.Delete(storageId, filename); err != nil {
			log.Println(err)
			_, _ = w.Write([]byte("Error del"))
		}
		_, _ = w.Write([]byte("Successful delete"))
	}
}

func (h *Handler) GetFile(w http.ResponseWriter, r *http.Request) {

	sessionId := r.Header.Get("Session")
	storageId := h.Sessions.GetIdStorage(sessionId)

	filename := r.PathValue("id")
	e, err := h.FilesStorage.SelectFile(context.Background(), storageId, filename)
	defer e.F.Close()
	if err != nil {
		_, _ = fmt.Fprint(w, "Get failed")
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+e.Filename)
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))

	_, err = io.Copy(w, e.F)
	if err != nil {
		log.Println("error send file to user: ", err)
	}

}

func (h *Handler) AddLinkForFile(w http.ResponseWriter, r *http.Request) {

	var l struct {
		Filename       string `json:"filename"`
		NumberOfVisits string `json:"numberOfVisits"`
		Lifetime       string `json:"lifetime"`
	}
	var s []byte

	s, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(s, &l)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	sessionId := r.Header.Get("Session")
	storageId := h.Sessions.GetIdStorage(sessionId)
	file, err := h.FiLiInfo.F.Select(storageId, l.Filename)
	if err != nil {
		log.Println("no such file: ", err)
	}
	link := uuid.NewString()
	var sqlVisits sql.NullInt16
	var sqlLifetime sql.NullTime

	if l.NumberOfVisits != "" {
		tmp, _ := strconv.Atoi(l.NumberOfVisits)
		sqlVisits.Int16 = int16(tmp)
		sqlVisits.Valid = true
	} else {
		sqlVisits.Valid = false
	}

	if l.Lifetime != "" {
		tmp, _ := strconv.Atoi(l.Lifetime)
		sqlLifetime.Time = time.Unix(time.Now().Unix()+int64(tmp), 0)
		sqlLifetime.Valid = true
	} else {
		sqlLifetime.Valid = false
	}

	if err := h.FiLiInfo.L.Insert(file_dao.Link{FileID: file.FileID, LinkID: link,
		NumberOfVisits: sqlVisits, Lifetime: sqlLifetime}); err != nil {
		log.Println("error insert link: ", err)
		_, _ = fmt.Fprint(w, "error")
	}
	ans, _ := json.Marshal(
		struct {
			Answer string
			Link   string
		}{
			Answer: "link created",
			Link:   "/link/" + link})
	_, _ = w.Write(ans)
}

func (h *Handler) GetFileFromLink(w http.ResponseWriter, r *http.Request) {
	linkID := r.PathValue("id")

	linkInfo, err := h.FiLiInfo.L.Select(linkID)
	if err != nil {
		log.Println("Error select file", err)
		_, _ = fmt.Fprint(w, "Link not found")
		return
	}
	if (linkInfo.NumberOfVisits.Int16 > 0 || !linkInfo.NumberOfVisits.Valid) &&
		(!linkInfo.Lifetime.Valid || linkInfo.Lifetime.Time.Unix()-time.Now().Unix() > 0) {
		file, err := h.FiLiInfo.F.SelectByID(linkInfo.FileID)
		if err != nil {
			log.Println("error select file:", err)
		}
		e, err := h.FilesStorage.SelectFile(context.Background(), file.StorageID, file.Filename)
		defer e.F.Close()
		if err != nil {
			_, _ = fmt.Fprint(w, "Get failed")
		} else {
			w.Header().Set("Content-Disposition", "attachment; filename="+e.Filename)
			w.Header().Set("Content-Type", "multipart/form-data")

			//http.ServeContent(w, r, e.Filename, time.Now(), e.F)
			_, err = io.Copy(w, e.F)
			if err != nil {
				log.Println("error send file to user: ", err)
			}
		}
	} else {
		_, _ = fmt.Fprint(w, "link unreachable")
		return
	}

	if linkInfo.NumberOfVisits.Valid && linkInfo.NumberOfVisits.Int16 > 0 {
		linkInfo.NumberOfVisits.Int16--
		err := h.FiLiInfo.L.Update(linkInfo)
		if err != nil {
			log.Println("error update link:", err)
		}
	}
}

func (h *Handler) DeleteLink(w http.ResponseWriter, r *http.Request) {
	var l struct {
		Filename string `json:"filename"` // возможно не нужно
		LinkID   string `json:"linkID"`
	}
	var s []byte

	s, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(s, &l)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	sessionId := r.Header.Get("Session")
	storageId := h.Sessions.GetIdStorage(sessionId)
	linkInfo, _ := h.FiLiInfo.L.Select(l.LinkID)
	file, err := h.FiLiInfo.F.SelectByID(linkInfo.FileID)
	if errors.Is(err, file_dao.ErrSelect) || file.StorageID != storageId {
		log.Println(err)
		fmt.Fprint(w, "Error deleting")
		return
	}

	if err := h.FiLiInfo.L.Delete(l.LinkID); err != nil {
		log.Println("error deleting link", err)
	}
	_, _ = fmt.Fprint(w, "link was delete")
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

func (h *Handler) Mux(mux *http.ServeMux) {
	mux.HandleFunc("GET /users", h.AuthMiddleware(h.ShowAll))
	mux.HandleFunc("GET /answer", h.GetAnswer)
	mux.HandleFunc("POST /user", h.Registration)
	mux.HandleFunc("GET /user", h.Login)
	mux.HandleFunc("POST /file", h.AuthMiddleware(h.UploadFile))
	mux.HandleFunc("DELETE /file/{id}", h.AuthMiddleware(h.DeleteFile))
	mux.HandleFunc("GET /file/{id}", h.AuthMiddleware(h.GetFile))
	mux.HandleFunc("POST /link", h.AuthMiddleware(h.AddLinkForFile))
	mux.HandleFunc("GET /link/{id}", h.GetFileFromLink)
	mux.HandleFunc("DELETE /link", h.AuthMiddleware(h.DeleteLink))
}
