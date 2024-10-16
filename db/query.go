package db

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (dbh *DbHandler) GetPricesCollection() *mongo.Collection {
	return dbh.Client.Database(dbh.conf.DBName).Collection(dbh.conf.PricesColletionName)
}

func (dbh *DbHandler) GetShopsCollection() *mongo.Collection {
	return dbh.Client.Database(dbh.conf.DBName).Collection(dbh.conf.ShopsColletionName)
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

func (dbh *DbHandler) CreateShop(l *logrus.Entry, shop *Shop) (*Shop, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := dbh.GetShopsCollection().InsertOne(ctx, shop)
	if err != nil {
		l.WithError(err).Error("Failed to insert shop")
		return nil, err
	}

	shop.ID = result.InsertedID.(primitive.ObjectID)
	return shop, nil
}

func (dbh *DbHandler) GetShops(l *logrus.Entry) (*[]Shop, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := dbh.GetShopsCollection().Find(ctx, bson.M{})
	if err != nil {
		l.WithError(err).Error("Failed to get shops")
		return nil, err
	}
	defer cursor.Close(ctx)

	shops := make([]Shop, 0)
	if err = cursor.All(ctx, &shops); err != nil {
		l.WithError(err).Error("Failed to decode shops")
		return nil, err
	}

	return &shops, nil
}

func (dbh *DbHandler) GetShop(l *logrus.Entry, id primitive.ObjectID) (*Shop, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var shop Shop
	err := dbh.GetShopsCollection().FindOne(ctx, bson.M{"_id": id}).Decode(&shop)
	if err != nil {
		l.WithError(err).Error("Failed to get shop")
		return nil, err
	}

	return &shop, nil
}

func (dbh *DbHandler) UpdateShop(l *logrus.Entry, shop *Shop) (*Shop, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := dbh.GetShopsCollection()
	filter := bson.M{"_id": shop.ID}
	updateDoc := bson.M{"$set": shop}

	result := collection.FindOneAndUpdate(ctx, filter, updateDoc, options.FindOneAndUpdate().SetReturnDocument(options.After))

	var updated Shop
	if err := result.Decode(&updated); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		l.WithError(err).Error("Failed to update inventory item")
		return nil, err
	}

	return &updated, nil

}

func (dbh *DbHandler) DeleteShop(l *logrus.Entry, id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := dbh.GetShopsCollection().DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		l.WithError(err).Error("Failed to delete shop")
		return err
	}

	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

// Price operations

func (dbh *DbHandler) CreatePrice(l *logrus.Entry, price *Price) (*Price, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	price.CreatedAt = now
	price.UpdatedAt = now

	result, err := dbh.GetPricesCollection().InsertOne(ctx, price)
	if err != nil {
		l.WithError(err).Error("Failed to insert ingredient price")
		return nil, err
	}

	price.ID = result.InsertedID.(primitive.ObjectID)
	return price, nil
}

func (dbh *DbHandler) UpdatePrice(l *logrus.Entry, update *Price) (*Price, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := dbh.GetPricesCollection()
	filter := bson.M{"_id": update.ID}
	updateFields := Price{
		Price:     update.Price,
		UpdatedAt: time.Now(),
		Devise:    update.Devise,
	}
	updateDoc := bson.M{"$set": updateFields}

	result := collection.FindOneAndUpdate(ctx, filter, updateDoc, options.FindOneAndUpdate().SetReturnDocument(options.After))

	var updated Price
	if err := result.Decode(&updated); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		l.WithError(err).Error("Failed to update ingredient price")
		return nil, err
	}

	return &updated, nil
}

func (dbh *DbHandler) GetPrices(l *logrus.Entry, filter bson.M) (*[]Price, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := dbh.GetPricesCollection().Find(ctx, filter)
	if err != nil {
		l.WithError(err).Error("Failed to get ingredient prices")
		return nil, err
	}
	defer cursor.Close(ctx)

	prices := make([]Price, 0)
	if err = cursor.All(ctx, &prices); err != nil {
		l.WithError(err).Error("Failed to decode ingredient prices")
		return nil, err
	}

	return &prices, nil
}
