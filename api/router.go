package api

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"service/storage"
)

type Handler struct {
	D storage.Database
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
	if err := h.D.Authentication(context.TODO(), u); err != nil {
		answer, _ := json.Marshal(err.Error())
		w.Write(answer)
	} else {
		answer, _ := json.Marshal("Successful login")
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

func (h *Handler) Mux(mux *http.ServeMux) {
	mux.HandleFunc("GET /users", h.ShowAll)
	mux.HandleFunc("GET /answer", h.GetAnswer)
	mux.HandleFunc("POST /registration", h.Registration)
	mux.HandleFunc("GET /login", h.Login)
}
