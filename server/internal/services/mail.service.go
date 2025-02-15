package services

import (
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/arpansaha13/pariksha/internal/api"
)

var MailService api.MailServiceClient
var MailServiceConn *grpc.ClientConn

var mailServerAddr = "localhost:4010"

func init() {
	var err error
	MailServiceConn, err = grpc.NewClient(mailServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("could not connect to mail service: %s", err)
	}

	MailService = api.NewMailServiceClient(MailServiceConn)
}
