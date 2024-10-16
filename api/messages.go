package api

import (
	"catalog/messages"
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func (api *ApiHandler) ConsumeAddPriceMessage(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping AddPrice message consumption")
			return
		default:
			api.consumeAddPriceMessage(ctx)
			time.Sleep(time.Second)
		}
	}
}

func (api *ApiHandler) consumeAddPriceMessage(ctx context.Context) {
	ctx, span := api.tracer.Start(ctx, "consumeAddPriceMessage")
	defer span.End()
	l := logger.WithField("context", "consumeAddPriceMessage")
	ch, err := messages.OpenChannel(api.amqp)
	if err != nil {
		return
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		messages.AddPriceCatalogQueueName, // name
		true,                              // durable
		false,                             // delete when unused
		false,                             // exclusive
		false,                             // no-wait
		nil,                               // arguments
	)

	if err != nil {
		logger.WithError(err).Errorf("Failed to declare queue %s", q.Name)
		return
	}

	msgs, err := ch.Consume(
		q.Name,    // queue
		"catalog", // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)

	l = l.WithFields(logrus.Fields{
		"queue": q.Name})

	if err != nil {
		logger.WithError(err).Error("Failed to register a consumer")
		return
	}

	l.Info("Started consuming messages")

	for {
		select {
		case <-ctx.Done():
			l.Info("Shutting down consumer")
			return
		case msg, ok := <-msgs:
			if !ok {
				l.Warn("Message channel closed")
				return
			}
			messageCtx, messageSpan := api.tracer.Start(ctx, "processAddPriceMessage")
			startTime := time.Now()

			retryCount := 0
			maxRetries := 3
			var processErr error
			for retryCount < maxRetries {
				processErr = api.processAddPriceMessage(messageCtx, l, msg)
				if processErr == nil {
					break
				}
				retryCount++
				l.WithError(processErr).WithField("retry", retryCount).Warn("Retrying message processing")
				time.Sleep(time.Second * time.Duration(retryCount)) // Exponential backoff
			}

			duration := time.Since(startTime)
			processStatus := "success"
			if processErr != nil {
				processStatus = "failure"
			}
			messageSpan.SetAttributes(
				attribute.Int("retries", retryCount),
				attribute.String("status", processStatus),
				attribute.Int64("duration_ms", duration.Milliseconds()),
			)
			messageSpan.End()

			if processErr != nil {
				l.WithError(processErr).Error("Failed to process message after max retries")
				// Send to dead-letter queue
				err := ch.Publish(
					"",                           // exchange
					messages.DeadLetterQueueName, // routing key
					false,                        // mandatory
					false,                        // immediate
					amqp.Publishing{
						ContentType: "application/json",
						Body:        msg.Body,
						Headers: amqp.Table{
							"x-original-queue": messages.AddPriceCatalogQueueName,
							"x-error":          processErr.Error(),
						},
					},
				)
				if err != nil {
					l.WithError(err).Error("Failed to send message to dead-letter queue")
				}
			}

			msg.Ack(false)
		}
	}
}

func (api *ApiHandler) processAddPriceMessage(ctx context.Context, l *logrus.Entry, msg amqp.Delivery) error {
	ctx, span := api.tracer.Start(ctx, "processAddPriceMessage")
	defer span.End()
	l = l.WithContext(ctx).WithField("function", "processMessageAddPrice")
	var price messages.AddPrice
	if err := json.Unmarshal(msg.Body, &price); err != nil {
		span.SetStatus(codes.Error, "Failed to unmarshal message")
		span.RecordError(err)
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	span.SetAttributes(
		attribute.String("price.productId", price.ProductID),
		attribute.String("price.shopId", price.ShopID),
		attribute.Float64("price.price", price.Price),
		attribute.String("price.devise", price.Devise),
		attribute.String("price.date", price.Date.GoString()),
	)

	if err := api.validation.Validate.Struct(price); err != nil {
		span.SetStatus(codes.Error, "Failed to validate message")
		span.RecordError(err)
		return fmt.Errorf("failed to validate message: %w", err)
	}

	// We retrieve the last updated price for the product and associated shop
	// If the price is the same, we update the updatedAt field
	// Otherwise, we insert a new price
	dbCtx, dbSpan := api.tracer.Start(ctx, "retrieveAndUpdatePrice")
	l = l.WithContext(dbCtx).WithFields(
		logrus.Fields{
			"shopId":    price.ShopID,
			"productId": price.ProductID,
			"price":     fmt.Sprintf("%v %v", price.Price, price.Devise),
		},
	)
	defer dbSpan.End()
	lastPrice, err := api.dbh.GetLastUpdatedPrice(l, price.ShopID, price.ProductID)

	if err != nil && err != mongo.ErrNoDocuments {
		dbSpan.RecordError(err)
		dbSpan.SetStatus(codes.Error, "Failed to get last updated price")
		return fmt.Errorf("failed to get last updated price: %w", err)
	}

	if lastPrice != nil {

		if lastPrice.Price == price.Price && lastPrice.Devise == price.Devise {
			l.Info("Price is the same, updating the price in the DB")
			lastPrice.UpdatedAt = price.Date
			_, err := api.dbh.UpdatePrice(l, lastPrice)

			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "Failed to update price")
				return fmt.Errorf("failed to update price: %w", err)
			}

			return nil
		}
	} else {
		l.Debug("No price found for the given shop and product")
	}

	dbPrice := messages.NewPrice(&price)
	l.WithFields(logrus.Fields{
		"shopId":    dbPrice.ShopID,
		"productId": dbPrice.ProductID,
		"price":     fmt.Sprintf("%v %v", dbPrice.Price, dbPrice.Devise),
	}).Info("Inserting new price")
	if _, err = api.dbh.CreatePrice(l, dbPrice); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to insert new price")
		return fmt.Errorf("failed to insert new price: %w", err)
	}
	return nil
}
