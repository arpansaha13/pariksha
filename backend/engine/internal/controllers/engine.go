package controllers

import (
	"context"

	"pariksha/common/pkg/proto"
	"pariksha/engine/internal/services"
)

type Engine struct {
	engineSvc *services.Engine
}

func NewEngine(engineSvc *services.Engine) *Engine {
	return &Engine{
		engineSvc: engineSvc,
	}
}

func (c *Engine) Run(ctx context.Context, req *proto.RunCodeRequest) (*proto.RunCodeResponse, error) {
	return c.engineSvc.Run(ctx, req)
}
