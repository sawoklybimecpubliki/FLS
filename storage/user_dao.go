package storage

import (
	"FLS/storage/jwt"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
	"golang.org/x/crypto/bcrypt"
	"log"
	"strconv"
	"time"
)

type UserDAO struct {
	C *mongo.Collection
}

type Config struct {
	Host            string
	Port            int
	TimeoutSeconds  int
	DBName          string
	CollectionUsers string
	Path            string
}

func NewClient(cfg Config) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSeconds)*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx,
		options.Client().ApplyURI("mongodb://"+cfg.Host+":"+strconv.Itoa(cfg.Port)))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (db *UserDAO) AddNewUser(ctx context.Context, u User) error {
	if err := db.findExistingUser(ctx, u.Login); err != nil {
		return err
	}
	u.Password, _ = u.hash()
	u.IdStorage = u.Login
	_, err := db.C.InsertOne(ctx, u)
	if err != nil {
		log.Println("error insert", err)
	}
	return err
}

func (db *UserDAO) findByLogin(ctx context.Context, login string) (*User, error) {
	filter := bson.D{{"login", login}}
	var user User
	err := db.C.FindOne(ctx, filter).Decode(&user)

	switch err {
	case nil:
		return &user, nil
	case mongo.ErrNoDocuments:
		return nil, bsoncore.ErrElementNotFound
	default:
		return nil, errors.New("default") //TODO придумать ошибку
	}
}

func (db *UserDAO) findExistingUser(ctx context.Context, login string) error {
	if _, err := db.findByLogin(ctx, login); err == nil {
		return errors.New("Uncorrect username")
	}
	return nil
}

func (db *UserDAO) Authentication(ctx context.Context, u User) (string, error) {
	userDB, _ := db.findByLogin(ctx, u.Login)

	if userDB == nil {
		return "", errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userDB.Password), []byte(u.Password)); err != nil {
		return "", errors.New("invalid password")
	}

	token, err := jwt.NewToken(jwt.UserJWT{u.Logi