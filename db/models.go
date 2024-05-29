package db

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Ingredient struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id" validate:"omitempty"`
	Name        string             `bson:"name" json:"name" validate:"required"`
	Description string             `bson:"description" json:"description" validate:"required"`
	Type        string             `bson:"type" json:"type" validate:"required,oneof=vegetable fruit meat fish dairy spice sugar cereals nuts other"`
}
