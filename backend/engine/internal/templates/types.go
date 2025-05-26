package templates

type TemplateFunc func(code string, testCases string) (string, error)
