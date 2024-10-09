package db

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var loger = logrus.WithFields(logrus.Fields{
	"context": "db/query",
})

func (dbh *DbHandler) NewID() primitive.ObjectID {
	return primitive.NewObjectID()
}

func (dbh *DbHandler) GetIngredientsCollection() *mongo.Collection {
	return dbh.Client.Database(dbh.conf.DBName).Collection(dbh.conf.IngredientsCollectionName)
}

func (dbh *DbHandler) FindByID(l *logrus.Entry, id string) (*Ingredient, error) {
	// TODO Change those hardcoded values
	collection := dbh.GetIngredientsCollection()
	objectID, _ := primitive.ObjectIDFromHex(id)
	filter := map[string]primitive.ObjectID{"_id": objectID}
	var ingredient Ingredient
	err := collection.FindOne(context.Background(), filter).Decode(&ingredient)
	if err != nil {
		l.WithError(err).Error("Error when trying to find ingredient by ID")
		return nil, err
	}
	return &ingredient, nil
}

func (dbh *DbHandler) FindAllIngredients(l *logrus.Entry) (*[]Ingredient, error) {
	collection := dbh.GetIngredientsCollection()
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		l.WithError(err).Error("Error when trying to find all recipes")
		return nil, err
	}
	var ingredients []Ingredient
	err = cursor.All(context.Background(), &ingredients)
	if err != nil {
		l.WithError(err).Error("Error when trying to decode all recipes")
		return nil, err
	}
	return &ingredients, nil
}

func (dbh *DbHandler) FindByName(l *logrus.Entry, name string) (*Ingredient, error) {
	collection := dbh.GetIngredientsCollection()
	filter := map[string]string{"name": name}
	var ingredient Ingredient
	err := collection.FindOne(context.Background(), filter).Decode(&ingredient)
	if err != nil {
		l.WithError(err).Error("Error when trying to find ingredient by name")
		return nil, err
	}
	return &ingredient, nil
}

func (dbh *DbHandler) FindByType(l *logrus.Entry, ingredientType string) (*[]Ingredient, error) {
	collection := dbh.GetIngredientsCollection()
	filter := map[string]string{"type": ingredientType}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		l.WithError(err).Error("Error when trying to find ingredient by type")
		return nil, err
	}
	var ingredients []Ingredient
	if err = cursor.All(context.Background(), &ingredients); err != nil {
		l.WithError(err).Error("Error when trying to decode ingredients")
		return nil, err
	}
	return &ingredients, nil
}

func (dbh *DbHandler) InsertOne(l *logrus.Entry, ingredient *Ingredient) error {
	collection := dbh.GetIngredientsCollection()
	_, err := collection.InsertOne(context.Background(), ingredient)
	if err != nil {
		l.WithError(err).Error("Error when trying to insert ingredient")
		return err
	}
	return nil
}

func (dbh *DbHandler) UpsertOne(l *logrus.Entry, ingredient *Ingredient) error {
	collection := dbh.GetIngredientsCollection()
	filter := map[string]primitive.ObjectID{"_id": ingredient.ID}
	update := map[string]Ingredient{"$set": *ingredient}
	res, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		l.WithError(err).Error("Error when trying to upsert ingredient")
		return err
	}
	if res.MatchedCount == 0 {
		err = errors.New("ID not found")
		l.WithError(err).Error("Error when trying to upsert ingredient")
		return err
	}
	return nil
}
