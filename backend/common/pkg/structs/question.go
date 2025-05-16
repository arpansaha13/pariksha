package structs

type MCQQuestion struct {
	Statement string   `json:"statement"`
	Options   []string `json:"options"`
}

type SubjectiveQuestion struct {
	Statement string `json:"statement"`
}
