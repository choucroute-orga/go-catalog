package api

type FindIngredientByTypeRequest struct {
	Type string `query:"type" json:"type" validate:"required,oneof=vegetable fruit meat fish dairy spice sugar cereals nuts other"`
}
