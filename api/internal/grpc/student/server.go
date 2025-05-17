package studentgrpc

import (
	"Elschool-API/internal/service"
	"context"
	"errors"
	apiv1 "github.com/Ilya-Repin/elschooler/protos/gen/api/api.v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Student interface {
	AddStudent(ctx context.Context, userToken, login, password string) (studentToken string, err error)
	DeleteStudent(ctx context.Context, userToken, studentToken string) (err error)
	UpdateStudent(ctx context.Context, userToken, studentToken, login, password string) (newStudentToken string, err error)
}

type serverAPI struct {
	apiv1.UnimplementedStudentServer
	student Student
}

func Register(gRPC *grpc.Server, student Student) {
	apiv1.RegisterStudentServer(gRPC, &serverAPI{student: student})
}

func (s *serverAPI) AddStudent(ctx context.Context, req *apiv1.AddStudentRequest) (*apiv1.AddStudentResponse, error) {
	if err := validateUUID4(req.GetUserToken(), "user token"); err != nil {
		return nil, err
	}

	if req.GetLogin() == "" {
		return nil, status.Error(codes.InvalidArgument, "login required")
	}

	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password required")
	}

	token, err := s.student.AddStudent(ctx, req.GetUserToken(), req.GetLogin(), req.GetPassword())

	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, "no such user")
		}

		return nil, status.Error(codes.Internal, "add student error")
	}

	return &apiv1.AddStudentResponse{StudentToken: token}, nil
}

func (s *serverAPI) DeleteStudent(ctx context.Context, req *apiv1.DeleteStudentRequest) (*apiv1.DeleteStudentResponse, error) {
	if err := validateUUID4(req.GetUserToken(), "user token"); err != nil {
		return nil, err
	}

	if req.GetStudentToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "student token required")
	}
	if parsedUUID, err := uuid.Parse(req.GetStudentToken()); err != nil || parsedUUID.Version() != uuid.Version(4) {
		return nil, status.Error(codes.InvalidArgument, "wrong student token format, uuid4 required")
	}

	err := s.student.DeleteStudent(ctx, req.GetUserToken(), req.GetStudentToken())

	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, "no such user")
		}

		if errors.Is(err, service.ErrStudentNotFound) {
			return nil, status.Error(codes.InvalidArgument, "no such student")
		}

		return nil, status.Error(codes.Internal, "delete user error")
	}

	return &apiv1.DeleteStudentResponse{Success: true}, nil
}

func (s *serverAPI) UpdateStudent(ctx context.Context, req *apiv1.UpdateStudentRequest) (*apiv1.UpdateStudentResponse, error) {
	if err := validateUUID4(req.GetUserToken(), "user token"); err != nil {
		return nil, err
	}
	if err := validateUUID4(req.GetStudentToken(), "student token"); err != nil {
		return nil, err
	}
	if req.GetLogin() == "" {
		return nil, status.Error(codes.InvalidArgument, "login required")
	}
	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password required")
	}

	token, err := s.student.UpdateStudent(ctx, req.GetUserToken(), req.GetStudentToken(), req.GetLogin(), req.GetPassword())

	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, "no such user")
		}

		if errors.Is(err, service.ErrStudentNotFound) {
			return nil, status.Error(codes.InvalidArgument, "no such student")
		}

		return nil, status.Error(codes.Internal, "update user error")
	}

	return &apiv1.UpdateStudentResponse{StudentToken: token}, nil
}

func validateUUID4(id, fieldName string) error {
	if id == "" {
		return status.Errorf(codes.InvalidArgument, "%s required", fieldName)
	}
	if parsedUUID, err := uuid.Parse(id); err != nil || parsedUUID.Version() != 4 {
		return status.Errorf(codes.InvalidArgument, "wrong %s format, uuid4 required", fieldName)
	}
	return nil
}
