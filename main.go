package main

import (
	"catalog/api"
	"catalog/configuration"
	"catalog/db"
	"catalog/messages"
	"catalog/validation"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

var logger = logrus.WithFields(logrus.Fields{
	"context": "main",
})

func main() {
	configuration.SetupLogging()
	logger.Info("Catalog API Starting...")

	conf := configuration.New()
	logger.Logger.SetLevel(conf.LogLevel)
	dbh, err := db.New(conf)
	if err != nil {
		return
	}

	val := validation.New(conf)
	r := api.New(val)
	v1 := r.Group(conf.ListenRoute)
	amqp := messages.New(conf)
	h := api.NewApiHandler(dbh, amqp, conf)

	tp := api.InitOtel()
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		cancel()
		if err := tp.Shutdown(ctx); err != nil {
			logger.WithError(err).Error("Error shutting down tracer provider")
		}
		if err := dbh.Client.Disconnect(context.TODO()); err != nil {
			logger.WithError(err).Error("Error closing mongo connection")
		}
		if err := amqp.Close(); err != nil {
			logger.WithError(err).Error("Error closing amqp connection")
		}
	}()

	go h.ConsumeAddPriceMessage(ctx)

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("Shutting down gracefully...")
		cancel()
	}()

	h.Register(v1)
	r.Logger.Fatal(r.Start(fmt.Sprintf("%v:%v", conf.ListenAddress, conf.ListenPort)))
}
