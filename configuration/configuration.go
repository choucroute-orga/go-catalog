package configuration

import (
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

var logger = logrus.WithFields(logrus.Fields{
	"context": "configuration/configuration",
})

type Configuration struct {
	ListenPort                string
	ListenAddress             string
	ListenRoute               string
	LogLevel                  logrus.Level
	EventStoreURI             string
	RabbitURI                 string
	DBURI                     string
	DBName                    string
	IngredientsCollectionName string
	ShopsCollectionName       string
	PricesColletionName       string
	TranslateValidation       bool
	JWTSecret                 string
	OtelServiceName           string
}

func New() *Configuration {

	conf := Configuration{}
	var err error

	logLevel := os.Getenv("LOG_LEVEL")
	if len(logLevel) < 1 || logLevel != "debug" && logLevel != "error" && logLevel != "info" && logLevel != "trace" && logLevel != "warn" {
		logrus.WithFields(logrus.Fields{
			"logLevel": logLevel,
		}).Info("logLevel not conform, use `info` ")
		conf.LogLevel = logrus.InfoLevel
	}

	if logLevel == "debug" {
		conf.LogLevel = logrus.DebugLevel
	} else if logLevel == "error" {
		conf.LogLevel = logrus.ErrorLevel
	} else if logLevel == "info" {
		conf.LogLevel = logrus.InfoLevel
	} else if logLevel == "trace" {
		conf.LogLevel = logrus.TraceLevel
	} else if logLevel == "warn" {
		conf.LogLevel = logrus.WarnLevel
	}

	conf.ListenPort = os.Getenv("API_PORT")
	conf.ListenAddress = os.Getenv("API_ADDRESS")
	conf.ListenRoute = os.Getenv("API_ROUTE")

	conf.EventStoreURI = os.Getenv("EVENTSTORE_URI")
	conf.DBURI = os.Getenv("MONGODB_URI")

	// Check if one of the event store or the DBURI is set
	if len(conf.EventStoreURI) < 1 && len(conf.DBURI) < 1 {
		logger.Error("EVENT_STORE_URI or MONGODB_URI is not set")
		os.Exit(1)
	}

	conf.RabbitURI = os.Getenv("RABBITMQ_URL")

	if len(conf.RabbitURI) < 1 {
		logger.Error("RABBITMQ_URL is not set")
		os.Exit(1)
	}

	// Extract the dbName from the DBURI if mongodb is used
	if len(conf.DBURI) > 1 {

		// Try to split the DBURI by "/" and get the 4th element
		splitedUri := strings.Split(conf.DBURI, "/")
		if len(splitedUri) < 4 {
			logger.Error("Failed to extract the DBName from the DBURI")
			os.Exit(1)
		}
		conf.DBName = splitedUri[3]
		logger.Debug("DBName: ", conf.DBName)
	}

	conf.IngredientsCollectionName = os.Getenv("MONGODB_INGREDIENTS_COLLECTION")
	conf.PricesColletionName = os.Getenv("MONGODB_PRICES_COLLECTION")
	conf.ShopsCollectionName = os.Getenv("MONGODB_SHOPS_COLLECTION")

	if len(conf.ShopsCollectionName) < 1 {
		logger.Error("MONGODB_SHOPS_COLLECTION is not set")
		os.Exit(1)
	}

	if len(conf.PricesColletionName) < 1 {
		logger.Error("MONGODB_PRICES_COLLECTION is not set")
		os.Exit(1)
	}

	if len(conf.IngredientsCollectionName) < 1 {
		logger.Error("MONGODB_INGREDIENTS_COLLECTION is not set")
		os.Exit(1)
	}

	conf.TranslateValidation, err = strconv.ParseBool(os.Getenv("TRANSLATE_VALIDATION"))

	if err != nil {
		logger.Error("Failed to parse bool for TRANSLATE_VALIDATION")
		os.Exit(1)
	}

	conf.JWTSecret = os.Getenv("JWT_SECRET")
	conf.OtelServiceName = os.Getenv("OTEL_SERVICE_NAME")
	return &conf
}
