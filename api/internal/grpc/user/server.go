package usergrpc

import (
	"context"
	apiv1 "github.com/Ilya-Repin/elschooler/protos/gen/api/api.v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type User interface {
	CreateUser(ctx context.Context, service string) (token string, err error)
}

type serverAPI struct {
	apiv1.UnimplementedUserServer
	user User
}

func Register(gRPC *grpc.Server, user User) {
	apiv1.RegisterUserServer(gRPC, &serverAPI{user: user})
}

func (s *serverAPI) RegUser(ctx context.Context, req *apiv1.RegUserRequest) (*apiv1.RegUserResponse, error) {
	if req.GetService() == "" {
		return nil, status.Error(codes.InvalidArgument, "service name required")
	}

	token, err := s.user.CreateUser(ctx, req.GetService())

	if err != nil {
		return nil, status.Error(codes.Internal, "create user error")
	}

	return &apiv1.RegUserResponse{UserToken: token}, nil
}
