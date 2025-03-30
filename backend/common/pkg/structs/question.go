package structs

type MCQQuestion struct {
	Statement string   `json:"statement"`
	Options   []string `json:"options"`
}

type GeneralQuestion struct {
	Statement string `json:"statement"`
}
