package structs

type MCQQuestion struct {
	Statement string   `json:"statement"`
	Options   []string `json:"options"`
}

type SubjectiveQuestion struct {
	Statement string `json:"statement"`
}

type CodingQuestionExample struct {
	Input       string  `json:"input"`
	Output      string  `json:"output"`
	Explanation *string `json:"explanation,omitempty"`
}

type CodingQuestion struct {
	Title     string                  `json:"title"`
	Statement string                  `json:"statement"`
	Examples  []CodingQuestionExample `json:"examples"`
}
