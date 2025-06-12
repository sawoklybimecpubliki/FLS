package user

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

type Store struct {
	c *mongo.Collection
}

func NewStore(collection *mongo.Collection) (*Store, error) {
	mod := mongo.IndexModel{
		Keys:    bson.M{"login": 1},
		Options: options.Index().SetUnique(true),
	}

	if _, err := collection.Indexes().CreateOne(context.TODO(), mod); err != nil {
		return nil, fmt.Errorf("could not create index: %w", err)
	}

	return &Store{
		c: collection,
	}, nil
}

type MongoConfig struct {
	Host     string
	Port     string
	Username string
	Password string
}

func NewMongoClient(ctx context.Context, cfg MongoConfig) (*mongo.Client, error) {
	uri := fmt.Sprintf("mongodb://%s:%s", cfg.Host, cfg.Port)

	clientOpts := options.Client().ApplyURI(uri).SetAuth(options.Credential{
		Username: cfg.Username,
		Password: cfg.Password,
	})

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("could not connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("could not ping MongoDB: %w", err)
	}

	return client, nil
}

func (db *Store) InsertUser(ctx context.Context, u User) error {
	log.Println("user: ", u)
	_, err := db.c.InsertOne(ctx, u)
	if err != nil {
		log.Println("error insert", err)
	}
	return err
}

func (db *Store) GetUser(ctx context.Context, u User) (User, error) {
	existingUser := User{}

	err := db.c.FindOne(ctx, bson.M{"login": u.Login}).Decode(&existingUser)
	if err != nil {
		return User{}, fmt.Errorf("could not find user: %w", err)
	}

	return existingUser, nil
}

func (db *Store) DeleteUser(ctx context.Context, u User) error {
	filter := bson.D{{
		Key:   "login",
		Value: u.Login,
	}}
	result, err := db.c.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return nil
	}
	return nil
}

/*func (db *Store) All(ctx context.Context) ([]User, error) {
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
}*/
