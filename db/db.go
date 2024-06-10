package db

import (
	"catalog/configuration"
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func New(conf *configuration.Configuration) (*mongo.Client, error) {

	// Database connexion

	loger.Info("Connecting to MongoDB..." + conf.DBURI)
	client, err := mongo.Connect(context.TODO(), options.Client().
		ApplyURI(conf.DBURI))
	if err != nil {
		panic(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		panic(err)
	}
	loger.Info("Connected to MongoDB!")
	return client, nil
}
