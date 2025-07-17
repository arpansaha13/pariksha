package test

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type TestCase[Req any, Resp any] struct {
	Name         string
	Metadata     metadata.MD
	ExpectedCode codes.Code
	GetRequest   func(map[string]any) Req
	Setup        func(*testing.T) map[string]any
	Validate     func(*testing.T, Resp, map[string]any)
}
