package storage

import (
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
	u.IdStorage = u.Login + "_storage"
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
		return nil, errors.New("User not found ")
	}
}

func (db *UserDAO) findExistingUser(ctx context.Context, login string) error {
	if _, err := db.findByLogin(ctx, login); err == nil {
		return errors.New("Uncorrect username")
	}
	return nil
}

func (db *UserDAO) Authentication(ctx context.Context, u User) error {
	userDB, _ := db.findByLogin(ctx, u.Login)

	if userDB == nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userDB.Password), []byte(u.Password)); err != nil {
		return errors.New("invalid password")
	}

	return nil
}

func (db *UserDAO) DeleteUser(ctx context.Context, id string) error {
	filter := bson.D{{"login", id}}
	result, err := db.C.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return nil
	}
	return nil
}

func (db *UserDAO) All(ctx context.Context) ([]User, error) {
	var u []User
	cur, err := db.C.Find(ctx, bson.D{})

	if err != nil {
		return nil, bsoncore.ErrElementNotFound
	}

	for cur.Next(ctx) {
		var elem User
		err := cur.Decode(&elem)
		if err != nil {
			log.Println("Decoding Failed")
		}
		u = append(u, elem)
	}

	if err := cur.Err(); err != nil {
		log.Println("cur: ", err)
	}

	return u, err
}
