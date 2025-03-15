package dtos

type UpdateCategoryDto struct {
	Name string `json:"name" validate:"required"`
}

type ReorderCategoryDto struct {
	Categories []int `json:"categories" validate:"required,min=1"`
}

type CategoryOrderDto struct {
	ID    int `json:"id" validate:"required"`
	Order int `json:"order" validate:"required"`
}
