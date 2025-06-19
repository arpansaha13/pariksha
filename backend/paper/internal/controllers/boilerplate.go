package controllers

import (
	"context"

	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/services"
)

type Boilerplate struct {
	service *services.Boilerplate
}

func NewBoilerplate(service *services.Boilerplate) *Boilerplate {
	return &Boilerplate{service: service}
}

// HandleGetBoilerplate handles fetching boilerplate code for a question
func (c *Boilerplate) HandleGetBoilerplate(ctx context.Context, req *proto.GetBoilerplateRequest) (*proto.GetBoilerplateResponse, error) {
	return c.service.GetBoilerplate(req.QuestionHash, req.LanguageId)
}
