package router

import (
	"encoding/json"
	"github.com/sawoklybimecpubliki/FLS/pkg/events"
	core "github.com/sawoklybimecpubliki/FLS/us/internal/core/user"
	"net/http"
)

func APIMux(handler *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", events.EventsMiddleware(handler.Register))
	mux.HandleFunc("POST /login", handler.Login)
	mux.HandleFunc("GET /logout", handler.Logout)
	mux.HandleFunc("GET /auth", handler.AuthCheck)

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

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var user User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		Respond("Invalid JSON", http.StatusBadRequest, w)
		return
	}

	err = h.app.Register(ctx, core.User{
		Login:    user.Username,
		Password: user.Password,
	})
	if err != nil {
		Respond(err.Error(), http.StatusBadRequest, w)
		return
	}
	Respond("OK", "success", w)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var user User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		Respond("Invalid JSON", http.StatusBadRequest, w)
		return
	}

	ses, err := h.app.Login(ctx, core.User{
		Login:    user.Username,
		Password: user.Password,
	})
	if err != nil {
		Respond(err.Error(), http.StatusBadRequest, w)
		return
	}

	SetAuthCookies(w, ses.SID, ses.Username)
	Respond("OK", "Success", w)
}

func (h *Handler) AuthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sid, username, err := GetAuthCookies(r)
	if err != nil {
		Respond(err.Error(), http.StatusBadRequest, w)
		return
	}

	err = h.app.CheckAuth(ctx, core.Session{SID: sid, Username: username})
	if err != nil {
		Respond(err.Error(), http.StatusBadRequest, w)
		return
	}

	SetAuthCookies(w, sid, username)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sid, username, err := GetAuthCookies(r)
	if err != nil {
		Respond(err.Error(), http.StatusBadRequest, w)
		return
	}

	err = h.app.Logout(ctx, core.Session{SID: sid, Username: username})
	if err != nil {
		Respond(err.Error(), http.StatusBadRequest, w)
		return
	}

	ClearAuthCookies(w)
	Respond("OK", "Success", w)
}
