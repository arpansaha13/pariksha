package controllers

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/services"
)

type TestCase struct {
	service *services.TestCase
}

func NewTestCase(service *services.TestCase) *TestCase {
	return &TestCase{service: service}
}

// HandleUpsertTestCases handles bulk creation and updates of test cases
func (c *TestCase) HandleUpsertTestCases(ctx context.Context, req *proto.UpsertTestCasesRequest) (*emptypb.Empty, error) {
	return c.service.UpsertTestCases(ctx, req)
}
