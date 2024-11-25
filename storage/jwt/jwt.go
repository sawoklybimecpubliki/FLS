package jwt

import (
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"log"
	"time"
)

const (
	signingKey = "storage"
)

type UserJWT struct {
	Name      string
	IdStorage uuid.UUID
}

func NewToken(user UserJWT, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := jwt.MapClaims{}
	claims["login"] = user.Name
	claims["storage"] = user.IdStorage
	claims["exp"] = time.Now().Add(duration).Unix()

	tokenString, err := token.SignedString([]byte(signingKey))
	if err != nil {
		log.Println("token: ", err)
		return "", err
	}
	return tokenString, nil
}
