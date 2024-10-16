package db

import (
	"catalog/configuration"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DbHandler struct {
	Client *mongo.Client
	conf   *configuration.Configuration
}

func NewDbHandler(client *mongo.Client, conf *configuration.Configuration) *DbHandler {
	handler := DbHandler{
		Client: client,
		conf:   conf,
	}
	return &handler
}

func New(conf *configuration.Configuration) (*DbHandler, error) {

	// Database connexion
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	loger.Info("Connecting to MongoDB..." + conf.DBURI)
	client, err := mongo.Connect(ctx, options.Client().
		ApplyURI(conf.DBURI))
	if err != nil {
		panic(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		panic(err)
	}
	loger.Info("Connected to MongoDB!")
	dbHandler := NewDbHandler(client, conf)
	return dbHandler, nil
}
