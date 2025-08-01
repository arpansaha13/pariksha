package modules

import (
	"context"
	"pariksha/common/pkg/proto"
	"pariksha/engine/internal/controllers"
)

type engineServer struct {
	proto.UnimplementedEngineServer

	engineCtrl *controllers.Engine
}

func (s *engineServer) RunCode(ctx context.Context, req *proto.RunCodeRequest) (*proto.RunCodeResponse, error) {
	return s.engineCtrl.Run(ctx, req)
}
