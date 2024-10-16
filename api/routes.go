package api

import (
	"catalog/db"
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var logger = logrus.WithField("context", "api/routes")

func (api *ApiHandler) getAliveStatus(c echo.Context) error {
	l := logger.WithField("request", "getAliveStatus")
	status := NewHealthResponse(LiveStatus)
	if err := c.Bind(status); err != nil {
		FailOnError(l, err, "Response binding failed")
		return NewInternalServerError(err)
	}
	l.WithFields(logrus.Fields{
		"action": "getStatus",
		"status": status,
	}).Debug("Health Status ping")

	return c.JSON(http.StatusOK, &status)
}

func (api *ApiHandler) getReadyStatus(c echo.Context) error {
	l := logger.WithField("request", "getReadyStatus")
	err := api.dbh.Client.Ping(context.Background(), nil)
	if err != nil {
		WarnOnError(l, err, "Unable to ping database to check connection.")
		return c.JSON(http.StatusServiceUnavailable, NewHealthResponse(NotReadyStatus))
	}

	return c.JSON(http.StatusOK, NewHealthResponse(ReadyStatus))
}

func (api *ApiHandler) postIngredient(c echo.Context) error {
	l := logger.WithField("request", "insertOne")

	// Retrieve the ingredient from the body
	ingredient := new(db.Ingredient)
	if err := c.Bind(ingredient); err != nil {
		FailOnError(l, err, "Request binding failed")
		return NewInternalServerError(err)
	}
	// Validate the ingredient
	if err := c.Validate(ingredient); err != nil {
		FailOnError(l, err, "Validation failed")
		return NewBadRequestError(err)
	}
	// Insert the ingredient
	ingredient.ID = api.dbh.NewID()

	i, _ := api.dbh.FindByName(l, ingredient.Name)
	if i != nil {
		return NewConflictError(errors.New("ingredient already exists"))
	}

	if err := api.dbh.InsertOne(l, ingredient); err != nil {
		FailOnError(l, err, "Insertion failed")
		return NewInternalServerError(err)
	}

	return c.JSON(http.StatusCreated, ingredient)

}

func (api *ApiHandler) getIngredients(c echo.Context) error {
	l := logger.WithField("request", "getIngredients")
	recipes, err := api.dbh.FindAllIngredients(l)
	if err != nil {
		return NewNotFoundError(err)
	}
	return c.JSON(http.StatusOK, recipes)
}

func (api *ApiHandler) getIngredientByID(c echo.Context) error {
	l := logger.WithField("request", "getIngredient")
	// Retrieve the ingredient ID from the path
	id := c.Param("id")

	// Find the ingredient
	ingredient, err := api.dbh.FindByID(l, id)
	if err != nil {
		return NewNotFoundError(err)
	}

	return c.JSON(http.StatusOK, ingredient)

}

func (api *ApiHandler) getIngredientByName(c echo.Context) error {
	l := logger.WithField("request", "getIngredient")
	// Retrieve the ingredient name from the path
	name := c.Param("name")
	// Find the ingredient
	ingredient, err := api.dbh.FindByName(l, name)
	if err != nil {
		WarnOnError(l, err, "Find by name failed")
		return NewNotFoundError(err)
	}

	return c.JSON(http.StatusOK, ingredient)
}

// Get the ingredient by the type in query parameter
func (api *ApiHandler) getIngredientByType(c echo.Context) error {
	l := logger.WithField("request", "getIngredientByType")

	// Retrieve the ingredient type from the query parameter
	ingredientType := c.Param("type")

	// Find the ingredient
	ingredients, err := api.dbh.FindByType(l, ingredientType)
	if err != nil {
		WarnOnError(l, err, "Find by type failed")
		return NewNotFoundError(err)
	}

	return c.JSON(http.StatusOK, ingredients)
}

func (api *ApiHandler) putIngredient(c echo.Context) error {
	l := logger.WithField("request", "putIngredient")
	// Retrieve the ingredient from the body
	ingredient := new(db.Ingredient)
	if err := c.Bind(ingredient); err != nil {
		FailOnError(l, err, "Request binding failed")
		return NewInternalServerError(err)
	}
	// Validate the ingredient
	if err := c.Validate(ingredient); err != nil {
		FailOnError(l, err, "Validation failed")
		return NewBadRequestError(err)
	}
	// Update the ingredient
	id, err := primitive.ObjectIDFromHex(c.Param("id"))

	if err != nil {
		WarnOnError(l, err, "Invalid ID")
		return NewBadRequestError(err)
	}
	ingredient.ID = id

	if err := api.dbh.UpsertOne(l, ingredient); err != nil {
		FailOnError(l, err, "Insertion failed")
		return NewNotFoundError(err)
	}

	return c.JSON(http.StatusOK, ingredient)
}

// Shop CRUD operations

func (api *ApiHandler) createShop(c echo.Context) error {
	ctx, span := api.tracer.Start(c.Request().Context(), "CreateShop")
	defer span.End()
	l := logger.WithContext(ctx).WithField("request", "CreateShop")

	var shop InsertShop
	if err := c.Bind(&shop); err != nil {
		return NewBadRequestError(err)
	}
	if err := c.Validate(shop); err != nil {
		return NewUnprocessableEntityError(err)
	}
	dbShop, err := NewInsertShop(&shop)

	if err != nil {
		return NewUnprocessableEntityError(err)
	}

	insertedShop, err := api.dbh.CreateShop(l, dbShop)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to insert shop")
		l.WithFields(
			logrus.Fields{
				"name":     shop.Name,
				"location": shop.Location,
				"id":       shop.ID,
			},
		).WithError(err).Error("Failed to insert shop")
		return NewInternalServerError(err)
	}

	l.WithFields(logrus.Fields{
		"id":       insertedShop.ID,
		"name":     insertedShop.Name,
		"location": insertedShop.Location,
	}).Debug("Shop created")

	return c.JSON(http.StatusCreated, insertedShop)
}

func (api *ApiHandler) getShop(c echo.Context) error {
	ctx, span := api.tracer.Start(c.Request().Context(), "GetShop")
	l := logger.WithContext(ctx).WithField("request", "GetShop")
	defer span.End()

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return NewBadRequestError(err)
	}

	shop, err := api.dbh.GetShop(l, id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return NewNotFoundError(err)
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get shop")
		l.WithError(err).Error("Failed to get shop")
		return NewInternalServerError(err)
	}

	return c.JSON(http.StatusOK, shop)
}

func (api *ApiHandler) getShops(c echo.Context) error {
	ctx, span := api.tracer.Start(c.Request().Context(), "GetShops")
	defer span.End()
	l := logger.WithContext(ctx).WithField("request", "GetShops")

	shops, err := api.dbh.GetShops(l)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get shops")
		l.WithError(err).Error("Failed to get shops")
		return NewInternalServerError(err)
	}

	return c.JSON(http.StatusOK, shops)
}

func (api *ApiHandler) updateShop(c echo.Context) error {
	ctx, span := api.tracer.Start(c.Request().Context(), "UpdateShop")
	defer span.End()
	l := logger.WithContext(ctx).WithField("request", "UpdateShop")

	var shop UpdateShop
	if err := c.Bind(&shop); err != nil {
		return NewBadRequestError(err)
	}

	if err := c.Validate(shop); err != nil {
		return NewUnprocessableEntityError(err)
	}

	shopDb, err := NewUpdateShop(&shop)
	if err != nil {
		return NewBadRequestError(err)
	}

	shopDb, err = api.dbh.UpdateShop(l, shopDb)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return NewNotFoundError(err)
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to update shop")
		l.WithError(err).Error("Failed to update shop")
		return NewInternalServerError(err)
	}

	return c.JSON(http.StatusOK, shopDb)
}

func (api *ApiHandler) deleteShop(c echo.Context) error {
	ctx, span := api.tracer.Start(c.Request().Context(), "DeleteShop")
	defer span.End()
	l := logger.WithContext(ctx).WithField("request", "DeleteShop")

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return NewBadRequestError(errors.New("Invalid ID"))
	}

	span.SetAttributes(attribute.String("shop_id", id.Hex()))
	if err := api.dbh.DeleteShop(l, id); err != nil {
		if err == mongo.ErrNoDocuments {
			return NewNotFoundError(errors.New("Shop not found"))
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to delete shop")
		l.WithError(err).Error("Failed to delete shop")
		return NewInternalServerError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// Price operations

func (api *ApiHandler) createPrice(c echo.Context) error {
	l := logger.WithField("request", "CreatePrice")

	var price InsertPrice
	if err := c.Bind(&price); err != nil {
		return NewBadRequestError(err)
	}

	if err := c.Validate(price); err != nil {
		return NewUnprocessableEntityError(err)
	}
	ingPrice := NewInsertPrice(&price)

	result, err := api.dbh.CreatePrice(l, ingPrice)
	if err != nil {
		l.WithError(err).Error("Failed to insert ingredient price")
		return NewInternalServerError(err)
	}

	return c.JSON(http.StatusCreated, result)
}

func (api *ApiHandler) getPrices(c echo.Context) error {
	ctx, span := api.tracer.Start(c.Request().Context(), "GetPrices")
	defer span.End()
	l := logger.WithContext(ctx).WithField("request", "GetPrices")

	filter := bson.M{}

	if id := c.QueryParam("id"); id != "" {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return NewBadRequestError(err)
		}
		filter["_id"] = oid
	}

	if minPrice := c.QueryParam("minPrice"); minPrice != "" {
		filter["price"] = bson.M{"$gte": minPrice}
	}
	if maxPrice := c.QueryParam("maxPrice"); maxPrice != "" {
		if _, ok := filter["price"]; !ok {
			filter["price"] = bson.M{}
		}
		filter["price"].(bson.M)["$lte"] = maxPrice
	}
	if ingID := c.QueryParam("ingId"); ingID != "" {
		filter["productId"] = ingID
	}
	if shopID := c.QueryParam("shopId"); shopID != "" {
		filter["shopId"] = shopID
	}
	if startDate := c.QueryParam("startDate"); startDate != "" {
		filter["createdAt"] = bson.M{"$gte": startDate}
	}
	if endDate := c.QueryParam("endDate"); endDate != "" {
		if _, ok := filter["updatedAt"]; !ok {
			filter["updatedAt"] = bson.M{}
		}
		filter["updatedAt"].(bson.M)["$lte"] = endDate
	}

	prices, err := api.dbh.GetPrices(l, filter)
	if err != nil {
		l.WithError(err).Error("Failed to get ingredient prices")
		return NewInternalServerError(err)
	}

	return c.JSON(http.StatusOK, prices)
}
