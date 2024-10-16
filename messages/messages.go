package messages

import (
	"catalog/configuration"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

var logger = logrus.WithFields(logrus.Fields{
	"context": "messages",
})

const (
	AddPriceCatalogQueueName = "catalog-add-price"
	DeadLetterQueueName      = "dead-letter-queue"
)

func New(conf *configuration.Configuration) *amqp.Connection {
	logger.Info("Connecting to RabbitMQ..." + conf.RabbitURI)
	conn, err := amqp.Dial(conf.RabbitURI)
	if err != nil {
		panic(err)
	}
	logger.Info("Connected to RabbitMQ!")
	return conn
}

func GetDeadLetterQueue(conn *amqp.Connection) *amqp.Queue {
	ch, err := OpenChannel(conn)
	if err != nil {
		logger.WithError(err).Error("Failed to open a channel")
		return nil
	}

	q, err := ch.QueueDeclare(
		DeadLetterQueueName, // name
		true,                // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		logger.WithError(err).Error("Failed to declare a queue")
	}

	return &q
}

func OpenChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		logger.WithError(err).Error("Failed to open a channel")
		return nil, err
	}
	return ch, nil
}
