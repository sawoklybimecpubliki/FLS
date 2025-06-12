package foundation

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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
