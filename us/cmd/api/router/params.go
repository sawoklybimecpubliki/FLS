package router

import (
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type User struct {
	Username string `json:"login"`
	Password string `json:"password"`
}

func SetAuthCookies(w http.ResponseWriter, sid uuid.UUID, username string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "sid",
		Value:    sid.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // true for HTTPS only
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "username",
		Value:    username,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})
}

func ClearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "sid",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // true for HTTPS only
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "username",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

func GetAuthCookies(r *http.Request) (uuid.UUID, string, error) {
	sidCookie, err := r.Cookie("sid")
	if err != nil {
		return uuid.UUID{}, "", fmt.Errorf("error getting sid cookie: %w", err)
	}

	sid, err := uuid.Parse(sidCookie.Value)
	if err != nil {
		return uuid.UUID{}, "", fmt.Errorf("error parsing sid cookie: %w", err)
	}

	usernameCookie, err := r.Cookie("username")
	if err != nil {
		return uuid.UUID{}, "", fmt.Errorf("error getting sid cookie: %w", err)
	}

	return sid, usernameCookie.Value, nil
}
