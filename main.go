package main

import (
	"catalog/api"
	"catalog/configuration"
	"catalog/db"
	"catalog/validation"
	"fmt"

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

	h := api.NewApiHandler(dbh, conf)

	h.Register(v1)
	r.Logger.Fatal(r.Start(fmt.Sprintf("%v:%v", conf.ListenAddress, conf.ListenPort)))
}
