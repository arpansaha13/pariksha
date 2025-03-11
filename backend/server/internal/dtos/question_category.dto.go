package dtos

type UpdateCategoryDto struct {
	Name string `json:"name" validate:"required"`
}

type ReorderCategoryDto struct {
	Categories []CategoryOrderDto `json:"categories" validate:"required,dive"`
}

type CategoryOrderDto struct {
	ID    int `json:"id" validate:"required"`
	Order int `json:"order" validate:"required"`
}
