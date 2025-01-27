package dtos

type CreatePaperDto struct {
	Title string `json:"title"`
}

type UpdatePaperDto struct {
	Title string `json:"title"`
}

type GetUserPapersResponse struct {
	ID             int                         `json:"id"`
	Title          string                      `json:"title"`
	MaxScore       int                         `json:"max_score"`
	PaperOwnership GetUserPapersPaperOwnership `json:"ownership"`
}

type GetUserPapersPaperOwnership struct {
	ID   int    `json:"id"`
	Path string `json:"path"`
	Type string `json:"type"`
}
