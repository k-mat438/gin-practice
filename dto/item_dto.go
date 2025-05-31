package dto

type CreateItemInput struct {
	Name        string `json:"name" binding:"required,min=2"`
	Price       uint   `json:"price" binding:"required,min=1,max=999999"`
	Description string `json:"description"`
}

type UpdateItemInput struct {
	Name        *string `json:"name" binding:"omitnill,min=2"`
	Price       *uint   `json:"price" binding:"omitenill,min=1,max=999999"`
	Description *string `json:"description"`
	SoldOut     *bool   `json:"soldOut"`
}
