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

// func LogAndReturnError(l *logrus.Entry, result *gorm.DB, action string, modelType string) error {
// 	if err := result.Error; err != nil {
// 		l.WithError(err).Error("Error when trying to query database to " + action + " " + modelType)
// 		return err
// 	}
// 	return nil
// }

func NewID() primitive.ObjectID {
	return primitive.NewObjectID()
}

func FindByID(l *logrus.Entry, mongo *mongo.Client, id string) (*Ingredient, error) {
	collection := mongo.Database("catalog").Collection("ingredient")
	filter := map[string]string{"_id": id}
	var ingredient Ingredient
	err := collection.FindOne(context.Background(), filter).Decode(&ingredient)
	if err != nil {
		l.WithError(err).Error("Error when trying to find ingredient by ID")
		return nil, err
	}
	return &ingredient, nil
}

func FindAllIngredients(l *logrus.Entry, mongo *mongo.Client) (*[]Ingredient, error) {
	collection := mongo.Database("catalog").Collection("ingredient")
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

func FindByName(l *logrus.Entry, mongo *mongo.Client, name string) (*Ingredient, error) {
	collection := mongo.Database("catalog").Collection("ingredient")
	filter := map[string]string{"name": name}
	var ingredient Ingredient
	err := collection.FindOne(context.Background(), filter).Decode(&ingredient)
	if err != nil {
		l.WithError(err).Error("Error when trying to find ingredient by name")
		return nil, err
	}
	return &ingredient, nil
}

func FindByType(l *logrus.Entry, mongo *mongo.Client, ingredientType string) (*[]Ingredient, error) {
	collection := mongo.Database("catalog").Collection("ingredient")
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

func InsertOne(l *logrus.Entry, mongo *mongo.Client, ingredient *Ingredient) error {
	collection := mongo.Database("catalog").Collection("ingredient")
	_, err := collection.InsertOne(context.Background(), ingredient)
	if err != nil {
		l.WithError(err).Error("Error when trying to insert ingredient")
		return err
	}
	return nil
}

func UpsertOne(l *logrus.Entry, mongo *mongo.Client, ingredient *Ingredient) error {
	collection := mongo.Database("catalog").Collection("ingredient")
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
