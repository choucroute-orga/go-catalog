// This handler uses both a MongoDB and the eventStore to store the prices.
package db

import (
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MixedHandler struct {
	mongoHandler *MongoHandler
	eventHandler *EventHandler
}

func NewMixedHandler(mongoHandler *MongoHandler, eventHandler *EventHandler) *MixedHandler {
	handler := MixedHandler{
		mongoHandler: mongoHandler,
		eventHandler: eventHandler,
	}
	return &handler
}

func (h *MixedHandler) Disconnect() error {
	err := h.mongoHandler.Disconnect()
	if err != nil {
		return err
	}
	return h.eventHandler.Disconnect()
}

func (h *MixedHandler) Ping() error {
	err := h.mongoHandler.Ping()
	if err != nil {
		return err
	}
	return h.eventHandler.Ping()
}

func (h *MixedHandler) NewID() primitive.ObjectID {
	return h.mongoHandler.NewID()
}

func (h *MixedHandler) FindByID(l *logrus.Entry, id string) (*Ingredient, error) {
	return h.mongoHandler.FindByID(l, id)
}

func (h *MixedHandler) FindAllIngredients(l *logrus.Entry) (*[]Ingredient, error) {
	return h.mongoHandler.FindAllIngredients(l)
}

func (h *MixedHandler) FindByName(l *logrus.Entry, name string) (*Ingredient, error) {
	return h.mongoHandler.FindByName(l, name)
}

func (h *MixedHandler) FindByType(l *logrus.Entry, ingredientType string) (*[]Ingredient, error) {
	return h.mongoHandler.FindByType(l, ingredientType)
}

func (h *MixedHandler) InsertOne(l *logrus.Entry, ingredient *Ingredient) error {
	return h.mongoHandler.InsertOne(l, ingredient)
}

func (h *MixedHandler) UpsertOne(l *logrus.Entry, ingredient *Ingredient) error {
	return h.mongoHandler.UpsertOne(l, ingredient)
}

func (h *MixedHandler) CreateShop(l *logrus.Entry, shop *Shop) (*Shop, error) {
	return h.mongoHandler.CreateShop(l, shop)
}

func (h *MixedHandler) GetShops(l *logrus.Entry) (*[]Shop, error) {
	return h.mongoHandler.GetShops(l)
}

func (h *MixedHandler) GetShop(l *logrus.Entry, id primitive.ObjectID) (*Shop, error) {
	return h.mongoHandler.GetShop(l, id)
}

func (h *MixedHandler) UpdateShop(l *logrus.Entry, shop *Shop) (*Shop, error) {
	return h.mongoHandler.UpdateShop(l, shop)
}

func (h *MixedHandler) DeleteShop(l *logrus.Entry, id primitive.ObjectID) error {
	return h.mongoHandler.DeleteShop(l, id)
}

func (h *MixedHandler) CreatePrice(l *logrus.Entry, price *Price) (*Price, error) {
	return h.eventHandler.CreatePrice(l, price)
}

func (h *MixedHandler) UpdatePrice(l *logrus.Entry, price *Price) (*Price, error) {
	return h.eventHandler.UpdatePrice(l, price)
}

func (h *MixedHandler) GetPrices(l *logrus.Entry) (*[]Price, error) {
	return h.eventHandler.GetPrices(l)
}

func (h *MixedHandler) GetLastUpdatedPrice(l *logrus.Entry, shopID, productID string) (*Price, error) {
	return h.eventHandler.GetLastUpdatedPrice(l, shopID, productID)
}