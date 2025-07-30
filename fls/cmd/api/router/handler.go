package router

import (
	"encoding/json"
	"errors"
	"github.com/sawoklybimecpubliki/FLS/fls/internal/core"
	"io"
	"log"
	"mime/multipart"
	"net/http"
)

func APIMux(handler *Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /file/upload", handler.UploadFile)
	mux.HandleFunc("DELETE /file/del", handler.DeleteFile)
	mux.HandleFunc("GET /file", handler.GetFile)
	mux.HandleFunc("POST /link/create", handler.AddLinkForFile)
	mux.HandleFunc("GET /link", handler.GetFileFromLink)
	mux.HandleFunc("DELETE /link/delete", handler.DeleteLink)

	return mux
}

type Handler struct {
	app *core.Service
}

func NewHandler(service *core.Service) *Handler {
	return &Handler{
		app: service,
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
	cookie, _ := r.Cookie("username")
	storageId := cookie.Value + "_storage"

	if err := h.app.UploadFile(r.Context(), core.Element{
		Filename: header.Filename,
		Size:     header.Size,
		F:        file,
	}, storageId); err != nil {
		log.Println(err)
		Respond(err, "error", w)
		return
	}

	Respond("File uploaded successfully", "Success", w)
}

func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {

	filename := r.URL.Query().Get("id")
	cookie, _ := r.Cookie("username")
	storageId := cookie.Value + "_storage"

	if err := h.app.DeleteFile(r.Context(), storageId, filename); err != nil {
		Respond("Error", http.StatusBadRequest, w)
		return
	}
	Respond("Successful delete", "Success", w)

}

func (h *Handler) GetFile(w http.ResponseWriter, r *http.Request) {

	cookie, _ := r.Cookie("username")
	storageId := cookie.Value + "_storage"

	filename := r.URL.Query().Get("id")
	e, err := h.app.Storage.SelectFile(r.Context(), storageId, filename)
	if err != nil {
		log.Println("Get failed")
		Respond(http.StatusBadRequest, errors.New("error"), w)
		return
	}
	defer func(F multipart.File) {
		err := F.Close()
		if err != nil {
			log.Println("Error close file", err)
		}
	}(e.F)

	w.Header().Set("Content-Disposition", "attachment; filename="+e.Filename)
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))

	_, err = io.Copy(w, e.F)
	if err != nil {
		log.Println("error send file to user: ", err)
		Respond(http.StatusBadRequest, errors.New("error"), w)
	}

}

func (h *Handler) AddLinkForFile(w http.ResponseWriter, r *http.Request) {

	var l core.LinkJSON
	var s []byte

	s, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(s, &l)
	if err != nil {
		Respond(http.StatusBadRequest, errors.New("error"), w)
		return
	}
	cookie, _ := r.Cookie("username")
	storageId := cookie.Value + "_storage"

	link, err := h.app.AddLink(storageId, l)
	if err != nil {
		Respond(http.StatusBadRequest, err, w)
		return
	}

	Respond("/link/"+link, "Success", w)
}

func (h *Handler) GetFileFromLink(w http.ResponseWriter, r *http.Request) {
	linkID := r.URL.Query().Get("id")

	e, err := h.app.GetFile(r.Context(), linkID)
	if err != nil {
		log.Println("Error get", err)
		Respond(http.StatusBadRequest, "No such file", w)
		return
	}
	defer func(F multipart.File) {
		err := F.Close()
		if err != nil {
			log.Println("Error close file", err)
		}
	}(e.F)

	w.Header().Set("Content-Disposition", "attachment; filename="+e.Filename)
	w.Header().Set("Content-Type", "multipart/form-data")
	_, err = io.Copy(w, e.F)
	if err != nil {
		log.Println("error send file to user: ", err)
		Respond(http.StatusBadRequest, errors.New("error send file"), w)
		return
	}
}

func (h *Handler) DeleteLink(w http.ResponseWriter, r *http.Request) {
	var l core.LinkJSON
	var s []byte

	s, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(s, &l)
	if err != nil {
		Respond(http.StatusBadRequest, errors.New("error"), w)
		return
	}

	cookie, _ := r.Cookie("username")
	storageId := cookie.Value + "_storage"

	if err := h.app.DeleteLink(r.Context(), storageId, l); err != nil {
		Respond(http.StatusBadRequest, err, w)
		return
	}

	Respond("link was delete", "Success", w)
}
