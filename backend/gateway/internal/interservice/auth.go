package interservice

import (
	"context"
	"fmt"
	"log"
	"sync"

	backoff "github.com/cenkalti/backoff/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/proto"
	"pariksha/gateway/internal/config/env"
)

var (
	authSvc     *authService
	authSvcOnce sync.Once
)

type authService struct {
	client proto.AuthClient
	conn   *grpc.ClientConn
}

func CloseAuthConn() {
	if authSvc != nil && authSvc.conn != nil {
		authSvc.conn.Close()
	}
}

func ensureAuthService() {
	authSvcOnce.Do(func() {
		authSvc = &authService{}
		addr := fmt.Sprintf("%s:%s", env.AUTH_SERVER_HOST, env.AUTH_SERVER_PORT)

		operation := func() (*grpc.ClientConn, error) {
			conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return nil, err
			}
			return conn, nil
		}

		// Use exponential backoff to retry connection
		ctx := context.Background()
		expBackoff := backoff.NewExponentialBackOff()
		conn, err := backoff.Retry(ctx, operation, backoff.WithBackOff(expBackoff))
		if err != nil {
			log.Fatalf("Failed to connect to auth service after retries: %v", err)
		}

		authSvc.conn = conn
		authSvc.client = proto.NewAuthClient(conn)
	})
}

var Authenticate = authenticate

func authenticate(req *proto.AuthenticateRequest) (*proto.AuthenticateResponse, error) {
	ensureAuthService()
	return authSvc.client.Authenticate(context.Background(), req)
}

var LoginWithPassword = loginWithPassword

func loginWithPassword(req *proto.LoginWithPasswordRequest) (*proto.UserResponse, *metadata.MD, error) {
	ensureAuthService()

	var header metadata.MD
	response, err := authSvc.client.LoginWithPassword(context.Background(), req, grpc.Header(&header))

	return response, &header, err
}

var SignUp = signUp

func signUp(req *proto.SignUpRequest) (*emptypb.Empty, error) {
	ensureAuthService()
	return authSvc.client.SignUp(context.Background(), req)
}

var VerifySignUp = verifySignUp

func verifySignUp(req *proto.VerificationRequest) (*proto.UserResponse, *metadata.MD, error) {
	ensureAuthService()

	var header metadata.MD
	response, err := authSvc.client.VerifySignUp(
		context.Background(),
		req,
		grpc.Header(&header),
	)

	return response, &header, err
}

var InitiateLoginWithOtp = initiateLoginWithOtp

func initiateLoginWithOtp(req *proto.LoginWithOtpRequest) (*emptypb.Empty, error) {
	ensureAuthService()
	return authSvc.client.InitiateLoginWithOtp(context.Background(), req)
}

var VerifyLoginOtp = verifyLoginOtp

func verifyLoginOtp(req *proto.VerificationRequest) (*proto.UserResponse, *metadata.MD, error) {
	ensureAuthService()

	var header metadata.MD
	response, err := authSvc.client.VerifyLoginOtp(
		context.Background(),
		req,
		grpc.Header(&header),
	)

	return response, &header, err
}

var ForgotPassword = forgotPassword

func forgotPassword(req *proto.ForgotPasswordRequest) (*emptypb.Empty, error) {
	ensureAuthService()
	return authSvc.client.ForgotPassword(context.Background(), req)
}

var ResetPassword = resetPassword

func resetPassword(req *proto.ResetPasswordRequest) (*emptypb.Empty, error) {
	ensureAuthService()
	return authSvc.client.ResetPassword(context.Background(), req)
}

var Logout = logout

func logout(req *proto.LogoutRequest) (*emptypb.Empty, error) {
	ensureAuthService()
	return authSvc.client.Logout(context.Background(), req)
}

var GetUser = getUser

func getUser(req *proto.GetUserRequest) (*proto.UserProfileResponse, error) {
	ensureAuthService()
	return authSvc.client.GetUser(context.Background(), req)
}

var UpsertUser = upsertUser

func upsertUser(req *proto.UpsertUserRequest) (*proto.UserProfileResponse, error) {
	ensureAuthService()
	return authSvc.client.UpsertUser(context.Background(), req)
}

var UpdateUser = updateUser

func updateUser(req *proto.UpdateUserRequest) (*proto.UpdateUserResponse, error) {
	ensureAuthService()
	return authSvc.client.UpdateUser(context.Background(), req)
}

func init() {
	ensureAuthService()
}
