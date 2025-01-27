package dtos

type CreatePaperDto struct {
	Title string `json:"title"`
}

type UpdatePaperDto struct {
	Title string `json:"title"`
}

type PaperResponse struct {
	ID             int                    `json:"id"`
	Title          string                 `json:"title"`
	MaxScore       int                    `json:"max_score"`
	PaperOwnership PaperOwnershipResponse `json:"ownership"`
}

type PaperOwnershipResponse struct {
	ID   int    `json:"id"`
	Path string `json:"path"`
	Type string `json:"type"`
}

type UpdatePaperResponse struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	MaxScore int    `json:"max_score"`
}
