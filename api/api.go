package api

import (
	"catalog/configuration"
	"catalog/db"

	"github.com/labstack/echo/v4"
)

type ApiHandler struct {
	dbh  *db.DbHandler
	conf *configuration.Configuration
}

func NewApiHandler(dbh *db.DbHandler, conf *configuration.Configuration) *ApiHandler {
	handler := ApiHandler{
		dbh:  dbh,
		conf: conf,
	}
	return &handler
}

func (api *ApiHandler) Register(v1 *echo.Group, conf *configuration.Configuration) {

	health := v1.Group("/health")
	health.GET("/alive", api.getAliveStatus)
	health.GET("/live", api.getAliveStatus)
	health.GET("/ready", api.getReadyStatus)

	ingredient := v1.Group("/ingredient")
	ingredient.POST("", api.postIngredient)
	ingredient.GET("", api.getIngredients)
	ingredient.PUT("/:id", api.putIngredient)
	ingredient.GET("/:id", api.getIngredientByID)
	ingredient.GET("/type/:type", api.getIngredientByType)
	ingredient.GET("/name/:name", api.getIngredientByName)
}
