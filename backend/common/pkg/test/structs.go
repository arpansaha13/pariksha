package test

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type TestCase[Req any, Resp any, SetupReturn any] struct {
	Name         string
	Metadata     metadata.MD
	ExpectedCode codes.Code
	GetRequest   func(SetupReturn) Req
	Setup        func(*testing.T) SetupReturn
	Validate     func(*testing.T, Resp, SetupReturn)
}
