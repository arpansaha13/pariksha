package interservice

import (
	"context"
	"fmt"
	"log"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/workers/exam/internal/config/env"
)

var (
	pSvc     *paperService
	pSvcOnce sync.Once
)

type paperService struct {
	client proto.PaperClient
	conn   *grpc.ClientConn
	ctx    context.Context
}

func ClosePaperConn() {
	if pSvc != nil && pSvc.conn != nil {
		pSvc.conn.Close()
	}
}

func ensurePaperService() {
	pSvcOnce.Do(func() {
		pSvc = &paperService{}
		addr := fmt.Sprintf("%s:%s", env.PAPER_SERVER_HOST, env.PAPER_SERVER_PORT)
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect to paper service: %v", err)
		}

		pSvc.conn = conn
		pSvc.client = proto.NewPaperClient(conn)

		md := metadata.New(map[string]string{
			constants.X_EXAM_API_TOKEN: env.EXAM_API_TOKEN,
		})
		pSvc.ctx = metadata.NewOutgoingContext(context.Background(), md)
	})
}

func init() {
	ensurePaperService()
}

var GetPaperQuestionsMeta = getPaperQuestionsMeta

func getPaperQuestionsMeta(paperHash string) ([]*proto.PaperQuestionMeta, error) {
	ensurePaperService()

	resp, err := pSvc.client.GetPaperQuestionsMeta(pSvc.ctx, &proto.PaperRequest{
		PaperHash: paperHash,
	})
	if err != nil {
		return nil, err
	}

	return resp.Questions, nil
}

var GetPaperCategoriesMeta = getPaperCategoriesMeta

func getPaperCategoriesMeta(paperHash string) ([]*proto.PaperCategoryMeta, error) {
	ensurePaperService()

	resp, err := pSvc.client.GetPaperCategoriesMeta(pSvc.ctx, &proto.PaperRequest{
		PaperHash: paperHash,
	})

	return resp.Categories, err
}
