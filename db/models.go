package db

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Ingredient struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id" validate:"omitempty"`
	Name     string             `bson:"name" json:"name" validate:"required"`
	ImageURL string             `bson:"image_url" json:"image_url" validate:"required"`
	Type     string             `bson:"type" json:"type" validate:"required,oneof=vegetable fruit meat fish dairy spice sugar cereals nuts other"`
}

type Price struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id" validate:"omitempty"`
	ProductID string             `bson:"productId" json:"productId" validate:"required"`
	ShopID    string             `bson:"shopId" json:"shopId" validate:"required"`
	Price     float64            `bson:"price" json:"price" validate:"required"`
	Devise    string             `bson:"devise" json:"devise" validate:"required,oneof=EUR USD"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt" validate:"required"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt" validate:"required"`
}

type Location struct {
	Street     string `bson:"street" json:"street" validate:"required"`
	PostalCode string `bson:"postal_code" json:"postal_code" validate:"required"`
	Country    string `bson:"country" json:"country" validate:"required"`
	City       string `bson:"city" json:"city" validate:"required"`
}

type Shop struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id" validate:"omitempty"`
	Name     string             `bson:"name" json:"name" validate:"required"`
	Location Location           `bson:"location" json:"location" validate:"required"`
}
