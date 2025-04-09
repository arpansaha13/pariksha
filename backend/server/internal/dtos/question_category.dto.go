package dtos

type UpdateCategoryDto struct {
	Name string `json:"name" validate:"required"`
}

type ReorderCategoryDto struct {
	Categories []int64 `json:"categories" validate:"required,min=1"`
}

type CategoryOrderDto struct {
	ID    int64 `json:"id" validate:"required"`
	Order int   `json:"order" validate:"required"`
}
