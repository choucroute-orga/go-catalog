package db

import (
	"catalog/configuration"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func New(conf *configuration.Configuration) (*mongo.Client, error) {

	// Database connexion

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s", conf.DBUser, conf.DBPassword, conf.DBHost, conf.DBPort, conf.DBName)
	loger.Info("Connecting to MongoDB..." + uri)
	client, err := mongo.Connect(context.TODO(), options.Client().
		ApplyURI(uri))
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
