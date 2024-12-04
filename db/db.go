package db

import (
	"catalog/configuration"
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DbHandler interface {
	Disconnect() error
	Ping() error
	NewID() primitive.ObjectID
	FindByID(l *logrus.Entry, id string) (*Ingredient, error)
	FindAllIngredients(l *logrus.Entry) (*[]Ingredient, error)
	FindByName(l *logrus.Entry, name string) (*Ingredient, error)
	FindByType(l *logrus.Entry, ingredientType string) (*[]Ingredient, error)
	InsertOne(l *logrus.Entry, ingredient *Ingredient) error
	UpsertOne(l *logrus.Entry, ingredient *Ingredient) error
	CreateShop(l *logrus.Entry, shop *Shop) (*Shop, error)
	GetShops(l *logrus.Entry) (*[]Shop, error)
	GetShop(l *logrus.Entry, id primitive.ObjectID) (*Shop, error)
	UpdateShop(l *logrus.Entry, shop *Shop) (*Shop, error)
	DeleteShop(l *logrus.Entry, id primitive.ObjectID) error
	CreatePrice(l *logrus.Entry, price *Price) (*Price, error)
	UpdatePrice(l *logrus.Entry, price *Price) (*Price, error)
	GetPrices(l *logrus.Entry) (*[]Price, error)
	GetLastUpdatedPrice(l *logrus.Entry, shopID, productID string) (*Price, error)
}

type MongoHandler struct {
	client                    *mongo.Client
	dbName                    string
	ingredientsCollectionName string
	shopsCollectionName       string
	pricesCollectionName      string
}

func newMongoHandler(client *mongo.Client, dbName, ingredientsCollectionName, shopsCollectionName, pricesCollectionName string) *MongoHandler {

	handler := MongoHandler{
		client:                    client,
		dbName:                    dbName,
		ingredientsCollectionName: ingredientsCollectionName,
		shopsCollectionName:       shopsCollectionName,
		pricesCollectionName:      pricesCollectionName,
	}
	return &handler
}

func NewMongoHandler(conf *configuration.Configuration) (*MongoHandler, error) {

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
	dbHandler := newMongoHandler(client, conf.DBName, conf.IngredientsCollectionName, conf.ShopsCollectionName, conf.PricesColletionName)
	return dbHandler, nil
}
