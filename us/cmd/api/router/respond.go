package router

import (
	"encoding/json"
	"log"
	"net/http"
)

type Response struct {
	Data   any `json:"data"`
	Status any `json:"status"`
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
	_, _ = w.Write(out)
}
