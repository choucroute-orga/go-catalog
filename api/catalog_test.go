package api

import (
	"catalog/configuration"
	"catalog/db"
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	dockertest "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient         *mongo.Client
	mongoPool           *dockertest.Pool
	mongoResource       *dockertest.Resource
	once                sync.Once
	CollectionsToCreate = []string{"ingredient", "price", "shop"}
	DBName              = "catalog"
	DBUser              = "root"
	DBPassword          = "password"
	DBPort              = "27017"
	DBHost              = "localhost"
	DBUri               = fmt.Sprintf("mongodb://%s:%s@%s:%s/%s", DBUser, DBPassword, DBHost, DBPort, DBName)
)

func SeedDatabase(mongo *mongo.Client) {
	// Create the recipe database and collection
	recipeDB := mongo.Database(DBName)
	for _, collection := range CollectionsToCreate {
		err := recipeDB.RunCommand(context.Background(), bson.D{{"create", collection}}).Err()
		if err != nil {
			logger.Warnf("Failed to create collection %s: %v", collection, err)
		}
	}
}

// InitTestMongo initializes a single MongoDB instance for all tests
func InitTestMongo() (*mongo.Client, error) {
	var initErr error
	once.Do(func() {
		// Create a new pool
		pool, err := dockertest.NewPool("")
		if err != nil {
			initErr = fmt.Errorf("could not construct pool: %w", err)
			return
		}

		mongoPool = pool

		// Set a timeout for docker operations
		pool.MaxWait = time.Second * 30

		// Start MongoDB container
		resource, err := pool.RunWithOptions(&dockertest.RunOptions{
			Repository: "mongo",
			Tag:        "5.0",
			Env: []string{
				fmt.Sprintf("MONGO_INITDB_ROOT_USERNAME=%v", DBUser),
				fmt.Sprintf("MONGO_INITDB_ROOT_PASSWORD=%v", DBPassword),
			},
		}, func(config *docker.HostConfig) {
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
		})

		if err != nil {
			initErr = fmt.Errorf("could not start resource: %w", err)
			return
		}

		mongoResource = resource
		DBPort = resource.GetPort("27017/tcp")
		mongoUri := fmt.Sprintf("mongodb://%s:%s@%s:%s", DBUser, DBPassword, DBHost, DBPort)
		DBUri = mongoUri
		// Initialize mongo client
		logger.Info("Connecting to MongoDB: " + DBUri)
		// Retry connection with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				initErr = fmt.Errorf("timeout waiting for mongodb to be ready")
				return
			case <-ticker.C:
				client, err := mongo.Connect(
					context.Background(),
					options.Client().ApplyURI(mongoUri).SetConnectTimeout(2*time.Second),
				)
				if err != nil {
					continue
				}

				// Try to ping
				if err := client.Ping(context.Background(), nil); err != nil {
					_ = client.Disconnect(context.Background())
					continue
				}

				mongoClient = client
				return
			}
		}
	})

	return mongoClient, initErr
}

// CleanupDatabase removes all data from the test database
func CleanupDatabase(t *testing.T, client *mongo.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, collection := range CollectionsToCreate {
		err := client.Database(DBName).Collection(collection).Drop(ctx)
		if err != nil {
			t.Logf("Warning: Failed to cleanup collection %s: %v", collection, err)
		}
	}

	err := client.Database(DBName).Drop(ctx)

	// Remove the inventory collection

	if err != nil {
		t.Logf("Warning: Failed to cleanup database: %v", err)
	}
}

func setupTest(t *testing.T) (*ApiHandler, func()) {
	t.Helper()

	// Use existing MongoDB instance
	client := mongoClient
	if client == nil {
		t.Fatal("MongoDB client not initialized")
	}

	// Clean the database
	CleanupDatabase(t, client)

	// Initialize the database
	SeedDatabase(client)

	// Create API handler
	conf := &configuration.Configuration{
		ListenAddress:             "localhost",
		ListenPort:                "3000",
		LogLevel:                  logrus.DebugLevel,
		DBURI:                     DBUri,
		DBName:                    DBName,
		IngredientsCollectionName: "ingredient",
		PricesColletionName:       "price",
		ShopsCollectionName:       "shop",
	}

	t.Log("DBUri", DBUri)
	dbh, err := db.NewMongoHandler(conf)
	if err != nil {
		t.Fatalf("Failed to create DB handler: %v", err)
	}
	api := NewApiHandler(dbh, nil, conf)

	// Return cleanup function
	return api, func() {
		CleanupDatabase(t, client)
	}
}

func TestMain(m *testing.M) {
	// Setup
	client, err := InitTestMongo()
	if err != nil {
		log.Fatalf("Could not start MongoDB: %s", err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = client.Disconnect(ctx)
	}
	if mongoPool != nil && mongoResource != nil {
		_ = mongoPool.Purge(mongoResource)
	}

	os.Exit(code)
}

func TestDB(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "Insert a price and update it in the DB",
			test: func(t *testing.T) {
				api, cleanup := setupTest(t)
				defer cleanup()
				l := logrus.WithField("test", "Insert ingredient in the DB")
				price := &db.Price{
					Price:     10.0,
					Devise:    "EUR",
					ProductID: primitive.NewObjectID().Hex(),
					ShopID:    primitive.NewObjectID().Hex(),
				}
				priceInserted, err := api.dbh.CreatePrice(l, price)
				if err != nil {
					t.Fatalf("Failed to insert price: %v", err)
				}

				if priceInserted.Price != 10.0 || priceInserted.Devise != "EUR" {
					t.Fatalf("Price not inserted: %v", priceInserted)
				}

				if len(priceInserted.ProductID) != 24 || len(priceInserted.ShopID) != 24 {
					t.Fatalf("ProductID or ShopID not valid: %v", priceInserted)
				}

				priceInserted.Price = 20.0
				priceInserted.Devise = "USD"
				priceUpdated, err := api.dbh.UpdatePrice(l, priceInserted)
				if err != nil {
					t.Fatalf("Failed to update price: %v", err)
				}
				if priceUpdated.Price != 20.0 || priceUpdated.Devise != "USD" {
					t.Fatalf("Price not updated: %v", priceUpdated)
				}
				if priceUpdated.ProductID != price.ProductID || priceUpdated.ShopID != price.ShopID {
					t.Fatalf("ProductID or ShopID changed: PID: %v, SID: %v", priceUpdated.ProductID, priceUpdated.ShopID)
				}

				if len(priceUpdated.ProductID) != 24 || len(priceUpdated.ShopID) != 24 {
					t.Fatalf("ProductID or ShopID not valid: %v", priceUpdated)
				}

				if priceUpdated.ID.IsZero() {
					t.Fatalf("ID not set: %v", priceUpdated)
				}

				// Get the price and check the same value
				// Check here that we get the good price
				price, err = api.dbh.GetLastUpdatedPrice(l, priceUpdated.ShopID, priceUpdated.ProductID)
				if err != nil {
					t.Fatalf("Failed to get prices: %v", err)
				}

				if price.Price != 20.0 || price.Devise != "USD" {
					t.Fatalf("Price not updated: %v", price)
				}

				l.Debug("Test started")
				l.Debugf("API: %v", api)
			},
		},
		{
			name: "Find the latest price by date",
			test: func(t *testing.T) {
				api, cleanup := setupTest(t)
				defer cleanup()
				l := logrus.WithField("test", "Find the latest price by date")
				price := &db.Price{
					Price:     10.0,
					Devise:    "EUR",
					ProductID: primitive.NewObjectID().Hex(),
					ShopID:    primitive.NewObjectID().Hex(),
					UpdatedAt: time.Now().Add(-time.Hour),
				}

				productId, _ := primitive.ObjectIDFromHex(price.ProductID)
				shopId, _ := primitive.ObjectIDFromHex(price.ShopID)
				price2 := &db.Price{
					Price:     20.0,
					Devise:    "USD",
					ProductID: productId.Hex(),
					ShopID:    shopId.Hex(),
					UpdatedAt: time.Now(),
				}

				_, err := api.dbh.CreatePrice(l, price)
				if err != nil {
					t.Fatalf("Failed to insert price: %v", err)
				}
				_, err = api.dbh.CreatePrice(l, price2)
				if err != nil {
					t.Fatalf("Failed to insert price2: %v", err)
				}

				// Get the price by updatedDate

				getPrice, err := api.dbh.GetLastUpdatedPrice(l, shopId.Hex(), productId.Hex())

				if err != nil {
					t.Fatalf("Failed to get price: %v", err)
				}

				if getPrice.Price != 20.0 || getPrice.Devise != "USD" {
					t.Fatalf("Latest price doesn't correspond: %v", getPrice)
				}

			},
		},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.test(t)
		})
	}
}
