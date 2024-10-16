package api

import (
	"catalog/db"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FindIngredientByTypeRequest struct {
	Type string `query:"type" json:"type" validate:"required,oneof=vegetable fruit meat fish dairy spice sugar cereals nuts other"`
}

type InsertShop struct {
	ID       string `json:"id" validate:"omitempty"`
	Name     string `json:"name" validate:"required,min=3"`
	Location struct {
		Street     string `json:"street" validate:"required"`
		PostalCode string `json:"postalCode" validate:"required"`
		Country    string `json:"country" validate:"required"`
		City       string `json:"city" validate:"required"`
	} `json:"location" validate:"required"`
}

type UpdateShop struct {
	ID         string `param:"id" validate:"omitempty"`
	InsertShop `json:",inline"`
}

type InsertPrice struct {
	ProductID string  `json:"productId" validate:"required"`
	ShopID    string  `json:"shopId" validate:"required"`
	Price     float64 `json:"price" validate:"required"`
	Devise    string  `json:"devise" validate:"required,oneof=EUR USD"`
}

type UpdatePrice struct {
	ID          string `param:"id" validate:"required"`
	InsertPrice `json:",inline"`
}

func NewInsertPrice(price *InsertPrice) *db.Price {
	return &db.Price{
		ProductID: price.ProductID,
		ShopID:    price.ShopID,
		Price:     price.Price,
		Devise:    price.Devise,
	}
}

func NewUpdatePrice(price *UpdatePrice) (*db.Price, error) {
	id, err := primitive.ObjectIDFromHex(price.ID)
	if err != nil {
		return nil, err
	}
	return &db.Price{
		ID:        id,
		ProductID: price.ProductID,
		ShopID:    price.ShopID,
		Price:     price.Price,
		Devise:    price.Devise,
	}, nil
}

func NewInsertShop(shop *InsertShop) (*db.Shop, error) {
	var err error

	dbShop := db.Shop{
		Name: shop.Name,
		Location: db.Location{
			Street:     shop.Location.Street,
			PostalCode: shop.Location.PostalCode,
			Country:    shop.Location.Country,
			City:       shop.Location.City,
		},
	}

	if shop.ID == "" {
		return &dbShop, err
	}
	dbShop.ID, err = primitive.ObjectIDFromHex(shop.ID)
	return &dbShop, err

}

func NewUpdateShop(shop *UpdateShop) (*db.Shop, error) {
	id, err := primitive.ObjectIDFromHex(shop.ID)
	if err != nil {
		return nil, err
	}
	return &db.Shop{
		ID:   id,
		Name: shop.Name,
		Location: db.Location{
			Street:     shop.Location.Street,
			PostalCode: shop.Location.PostalCode,
			Country:    shop.Location.Country,
			City:       shop.Location.City,
		},
	}, nil
}
