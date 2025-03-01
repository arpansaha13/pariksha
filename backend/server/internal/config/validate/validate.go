package validate

import "gopkg.in/go-playground/validator.v8"

var Do *validator.Validate

func init() {
	config := &validator.Config{TagName: "validate"}
	Do = validator.New(config)
}
