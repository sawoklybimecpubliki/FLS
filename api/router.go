package api

import (
	"FLS/filestorage"
	"FLS/storage"
	"FLS/storage/session"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Handler struct {
	D storage.Database
	F filestorage.Service
	S session.Service
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

	if err := h.D.AddNewUser(context.Background(), u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		answer, _ := json.Marshal("New user was add in database")
		w.Write(answer)
	}
}

func (h *Handler) ShowAll(w http.ResponseWriter, r *http.Request) {
	var u []storage.User

	u, _ = h.D.All(context.Background())
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
	if token, err := h.D.Authentication(context.TODO(), u); err != nil {
		answer, _ := json.Marshal(err.Error())
		w.Write(answer)
	} else {
		answer, _ := json.Marshal("Successful login")
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

	if err := h.D.DeleteUser(context.Background(), u.Login); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		answer, _ := json.Marshal("User was delete from database")
		w.Write(answer)
	}

}

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {

	file, header, _ := r.FormFile("file")
	// TODO сделать придумать как проверять пользователя
	defer file.Close()
	err := h.F.UploadFile(context.Background(), filestorage.Element{header.Filename, header.Size, file}, "sawok")

	if err != nil {
		log.Println(err)
		answer, error := json.Marshal(err)

		if error != nil {
			log.Println("marshaling error", error)
		}

		w.Write(answer)
	} else {

		answer, err := json.Marshal("File uploaded successfully")

		if err != nil {
			log.Println("marshaling error", err)
		}

		w.Write(answer)
	}

}

func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	var s []byte
	s, err := io.ReadAll(r.Body)

	if err != nil {
		log.Fatal(err)
	}

	var e filestorage.Element
	err = json.Unmarshal(s, &e)

	if err != nil {
		log.Println("marshal error")
	}

	if err := h.F.DeleteFile(context.Background(), "sawok", e.Filename); err != nil {
		log.Println("Deleting file error", err)
	} else {
		w.Write([]byte("Successful delete"))
	}
}

func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionId := r.Header.Get("Session")
		f, err := h.S.CheckSession(sessionId)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		if !f {
			h.S.SessionRefresh(sessionId)
		}
		next.ServeHTTP(w, r)
	}
}

func (h *Handler) Mux(mux *http.ServeMux) {
	mux.HandleFunc("GET /users", h.ShowAll)
	mux.HandleFunc("GET /answer", h.GetAnswer)
	mux.HandleFunc("POST /registration", h.Registration)
	mux.HandleFunc("GET /login", h.Login)
	mux.HandleFunc("POST /upload", h.UploadFile)
	mux.HandleFunc("POST /delete", h.DeleteFile)

}
