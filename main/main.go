package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

type user struct {
	Login    string `json:"Login"`
	Password string `json:"Password"`
}

type dataBase struct {
	data map[string][]byte
}

func (d dataBase) add(u user) {
	//idUser := uuid.New()
	//d.data = make(map[uuid.UUID]user)

	d.data[u.Login] = []byte(u.Password)
}

func (d dataBase) sel() dataBase {
	return d
	/*if _, ok := d.data[id]; !ok {
		slog.Error("user not found")
		return user{}
	}
	return user{d.data[id].Login, d.data[id].Password}*/
}

func (d dataBase) del(id string) error {
	if _, ok := d.data[id]; !ok {
		return errors.New("user not found")
	}
	delete(d.data, id)
	return nil
}

func (d dataBase) all() []user {
	var u []user
	for login, password := range d.data {
		u = append(u, user{login, string(password)})
	}
	return u
}

type handler struct {
	D dataBase
}

func (h *handler) compareLogin(loginNewUser string) error {
	db := h.D.sel()
	for login := range db.data {
		if login == loginNewUser {
			return errors.New("Uncorrect username") // отдать пользователю ошибку, но не сообщать что есть такой пользователь
		}
	}
	return nil
}

func (h *handler) registration(w http.ResponseWriter, r *http.Request) {
	var u user
	var s []byte
	s, err := io.ReadAll(r.Body)

	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(s, &u)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if err := h.compareLogin(u.Login); err != nil {
		answer, _ := json.Marshal(err.Error())
		w.Write(answer)
	} else {
		h.D.add(u)

		answer, _ := json.Marshal("New user was add in database")
		w.Write(answer)
	}
}

func (h *handler) showAll(w http.ResponseWriter, r *http.Request) {
	var u []user

	u = h.D.all()

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

func (h *handler) getAnswer(w http.ResponseWriter, r *http.Request) {
	answer, err := json.Marshal("Vse rabotaet, vot otvet")

	if err != nil {
		log.Fatal(err)
	}

	w.Write(answer)
}

func main() {
	var hand = handler{dataBase{make(map[string][]byte)}}
	router := http.NewServeMux()
	router.HandleFunc("GET /users", hand.showAll)
	router.HandleFunc("GET /answer", hand.getAnswer)
	router.HandleFunc("POST /registration", hand.registration)

	log.Println("server start listening on :8080")
	err := http.ListenAndServe("localhost:8080", router)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
