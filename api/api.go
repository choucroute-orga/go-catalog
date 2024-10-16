package api

import (
	"catalog/configuration"
	"catalog/db"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type ApiHandler struct {
	dbh   *db.DbHandler
	conf  *configuration.Configuration
	trace trace.Tracer
}

func NewApiHandler(dbh *db.DbHandler, conf *configuration.Configuration) *ApiHandler {
	handler := ApiHandler{
		dbh:   dbh,
		conf:  conf,
		trace: otel.Tracer(conf.OtelServiceName),
	}
	return &handler
}

func (api *ApiHandler) Register(v1 *echo.Group) {

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

	shop := v1.Group("/shop")
	shop.POST("", api.createShop)
	shop.GET("", api.getShops)
	shop.GET("/:id", api.getShop)
	shop.PUT("/:id", api.updateShop)
	shop.DELETE("/:id", api.deleteShop)

	price := v1.Group("/price")
	price.POST("", api.createPrice)
	price.GET("", api.getPrices)
}
