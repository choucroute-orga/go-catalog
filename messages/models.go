package messages

import (
	"catalog/db"
	"time"
)

type AddPrice struct {
	ProductID string    `json:"productId" validate:"required"`
	ShopID    string    `json:"shopId" validate:"required"`
	Price     float64   `json:"price" validate:"required,min=0.01"`
	Devise    string    `json:"devise" validate:"required,oneof=EUR USD"`
	Date      time.Time `json:"date" validate:"required"`
}

func NewPrice(price *AddPrice) *db.Price {
	return &db.Price{
		ProductID: price.ProductID,
		ShopID:    price.ShopID,
		Price:     price.Price,
		Devise:    price.Devise,
		UpdatedAt: price.Date,
		CreatedAt: time.Now(),
	}
}
