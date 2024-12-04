package db

import (
	"catalog/configuration"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/EventStore/EventStore-Client-Go/v4/esdb"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EventHandler struct {
	db *esdb.Client
}

func NewEventHandler(conf *configuration.Configuration) (*EventHandler, error) {
	settings, err := esdb.ParseConnectionString(conf.EventStoreURI)

	if err != nil {
		panic(err)
	}

	db, err := esdb.NewClient(settings)
	if err != nil {
		panic(err)
	}

	return &EventHandler{
		db: db,
	}, nil
}

func (e *EventHandler) Disconnect() error {
	panic("not implemented")
}

func (e *EventHandler) Ping() error {
	panic("not implemented")
}

func (e *EventHandler) NewID() primitive.ObjectID {
	panic("not implemented")
}

func (e *EventHandler) FindByID(l *logrus.Entry, id string) (*Ingredient, error) {
	panic("not implemented")
}

func (e *EventHandler) FindAllIngredients(l *logrus.Entry) (*[]Ingredient, error) {
	panic("not implemented")
}

func (e *EventHandler) FindByName(l *logrus.Entry, name string) (*Ingredient, error) {
	panic("not implemented")
}

func (e *EventHandler) FindByType(l *logrus.Entry, ingredientType string) (*[]Ingredient, error) {
	panic("not implemented")
}

func (e *EventHandler) InsertOne(l *logrus.Entry, ingredient *Ingredient) error {
	panic("not implemented")
}

func (e *EventHandler) UpsertOne(l *logrus.Entry, ingredient *Ingredient) error {
	panic("not implemented")
}

func (e *EventHandler) CreateShop(l *logrus.Entry, shop *Shop) (*Shop, error) {
	panic("not implemented")
}

func (e *EventHandler) GetShops(l *logrus.Entry) (*[]Shop, error) {
	panic("not implemented")
}

func (e *EventHandler) GetShop(l *logrus.Entry, id primitive.ObjectID) (*Shop, error) {
	panic("not implemented")
}

func (e *EventHandler) UpdateShop(l *logrus.Entry, shop *Shop) (*Shop, error) {
	panic("not implemented")
}

func (e *EventHandler) DeleteShop(l *logrus.Entry, id primitive.ObjectID) error {
	panic("not implemented")
}

const (
	PriceCreatedEventType = "PriceCreated"
	PriceUpdatedEventType = "PriceUpdated"
)

func (e *EventHandler) CreatePrice(l *logrus.Entry, price *Price) (*Price, error) {
	// Set current timestamp
	price.CreatedAt = time.Now()
	price.UpdatedAt = time.Now()

	// Construct stream name using shopId and productId
	streamName := fmt.Sprintf("price-%s-%s", price.ShopID, price.ProductID)

	// Serialize price to JSON
	priceJSON, err := json.Marshal(price)
	if err != nil {
		l.Errorf("Failed to marshal price: %v", err)
		return nil, err
	}

	// Prepare event data
	eventData := esdb.EventData{
		ContentType: esdb.ContentTypeJson,
		EventType:   PriceCreatedEventType,
		Data:        priceJSON,
	}

	// Write event to EventStore
	_, err = e.db.AppendToStream(context.Background(), streamName, esdb.AppendToStreamOptions{}, eventData)
	if err != nil {
		l.Errorf("Failed to append price creation event: %v", err)
		return nil, err
	}

	return price, nil
}

func (e *EventHandler) UpdatePrice(l *logrus.Entry, price *Price) (*Price, error) {
	// Set current timestamp for update
	price.CreatedAt = time.Now()
	price.UpdatedAt = time.Now()

	// Construct stream name using shopId and productId
	streamName := fmt.Sprintf("price-%s-%s", price.ShopID, price.ProductID)

	// Serialize price to JSON
	priceJSON, err := json.Marshal(price)
	if err != nil {
		l.Errorf("Failed to marshal price: %v", err)
		return nil, err
	}

	// Prepare event data
	eventData := esdb.EventData{
		ContentType: esdb.ContentTypeJson,
		EventType:   PriceUpdatedEventType,
		Data:        priceJSON,
	}

	// Write event to EventStore
	_, err = e.db.AppendToStream(context.Background(), streamName, esdb.AppendToStreamOptions{}, eventData)
	if err != nil {
		l.Errorf("Failed to append price update event: %v", err)
		return nil, err
	}

	return price, nil
}

func (e *EventHandler) GetLastUpdatedPrice(l *logrus.Entry, shopID, productID string) (*Price, error) {
	// Construct stream name
	streamName := fmt.Sprintf("price-%s-%s", shopID, productID)

	// Read stream from end, limit to 1 event
	stream, err := e.db.ReadStream(context.Background(), streamName, esdb.ReadStreamOptions{
		From:      esdb.End{},
		Direction: esdb.Backwards,
	}, 1)

	if err != nil {
		l.Errorf("Failed to read price stream: %v", err)
		return nil, err
	}
	defer stream.Close()

	// Deserialize the latest event
	var latestPrice Price

	for {
		event, err := stream.Recv()

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			l.Errorf("Failed to read price event: %v", err)
			return nil, err
		}

		// Deserialize event data
		err = json.Unmarshal(event.Event.Data, &latestPrice)
		if err != nil {
			l.Errorf("Failed to unmarshal price event data: %v", err)
			return nil, err
		}
	}

	return &latestPrice, nil
}

func (e *EventHandler) GetPrices(l *logrus.Entry) (*[]Price, error) {

	// Use a map to track the latest price for each unique stream
	latestPrices := make(map[string]Price)

	// Retrive all streams and return the latest event from each stream
	stream, err := e.db.ReadAll(context.Background(), esdb.ReadAllOptions{
		Direction: esdb.Backwards,
		From:      esdb.End{},
	}, 100)

	if err != nil {
		l.Errorf("Failed to read all streams: %v", err)
		return nil, err
	}
	defer stream.Close()

	for {
		event, err := stream.Recv()

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			panic(err)
		}

		if strings.HasPrefix(event.OriginalEvent().EventType, "$") {
			continue
		}

		// Only process price-related events
		if event.OriginalEvent().EventType != PriceCreatedEventType &&
			event.OriginalEvent().EventType != PriceUpdatedEventType {
			continue
		}

		// Deserialize event data
		var price Price
		err = json.Unmarshal(event.Event.Data, &price)
		if err != nil {
			l.Errorf("Failed to unmarshal price event data: %v", err)
			continue
		}

		// Create a unique key for the stream
		streamKey := fmt.Sprintf("price-%s-%s", price.ShopID, price.ProductID)

		// Store only the first (latest) event for each unique stream
		if _, exists := latestPrices[streamKey]; !exists {
			latestPrices[streamKey] = price
		}
	}

	// Convert map to slice
	prices := make([]Price, 0, len(latestPrices))
	for _, price := range latestPrices {
		prices = append(prices, price)
	}

	return &prices, nil

}
