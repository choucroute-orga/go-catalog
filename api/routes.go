package api

import (
	"catalog/db"
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	err := api.mongo.Ping(context.Background(), nil)
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
	ingredient.ID = db.NewID()

	i, _ := db.FindByName(l, api.mongo, ingredient.Name)
	if i != nil {
		return NewConflictError(errors.New("ingredient already exists"))
	}

	if err := db.InsertOne(l, api.mongo, ingredient); err != nil {
		FailOnError(l, err, "Insertion failed")
		return NewInternalServerError(err)
	}

	return c.JSON(http.StatusCreated, ingredient)

}

func (api *ApiHandler) getIngredients(c echo.Context) error {
	l := logger.WithField("request", "getIngredients")
	recipes, err := db.FindAllIngredients(l, api.mongo)
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
	ingredient, err := db.FindByID(l, api.mongo, id)
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
	ingredient, err := db.FindByName(l, api.mongo, name)
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
	ingredients, err := db.FindByType(l, api.mongo, ingredientType)
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

	if err := db.UpsertOne(l, api.mongo, ingredient); err != nil {
		FailOnError(l, err, "Insertion failed")
		return NewNotFoundError(err)
	}

	return c.JSON(http.StatusOK, ingredient)
}
