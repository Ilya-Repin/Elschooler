package marksgrpc

import (
	"Elschool-API/internal/domain/models"
	"Elschool-API/internal/service"
	"context"
	"errors"
	apiv1 "github.com/Ilya-Repin/elschooler/protos/gen/api/api.v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Marks interface {
	GetDayMarks(ctx context.Context, userToken, studentToken, date string) (marks models.DayMarks, err error)
	GetAverageMarks(ctx context.Context, userToken, studentToken string, period int32) (marks models.AverageMarks, err error)
	GetFinalMarks(ctx context.Context, userToken, studentToken string) (marks models.FinalMarks, err error)
}

const (
	emptyValue = 0
)

type serverAPI struct {
	apiv1.UnimplementedMarksServer
	marks Marks
}

func Register(gRPC *grpc.Server, marks Marks) {
	apiv1.RegisterMarksServer(gRPC, &serverAPI{marks: marks})
}

func (s *serverAPI) GetDayMarks(ctx context.Context, req *apiv1.DayMarksRequest) (*apiv1.DayMarksResponse, error) {
	if err := validateUUID4(req.GetUserToken(), "user token"); err != nil {
		return nil, err
	}
	if err := validateUUID4(req.GetStudentToken(), "student token"); err != nil {
		return nil, err
	}

	if req.GetDate() == "" {
		return nil, status.Error(codes.InvalidArgument, "date required")
	}

	dayMarks, err := s.marks.GetDayMarks(ctx, req.GetUserToken(), req.GetStudentToken(), req.GetDate())

	if err != nil {
		if errors.Is(err, service.ErrStudentNotFound) {
			return nil, status.Error(codes.InvalidArgument, "no such student")
		}

		return nil, status.Error(codes.Internal, "failed to get marks")
	}

	grpcMarks := make(map[string]*apiv1.LisOfIntMarks)

	for key, values := range dayMarks.Marks {
		grpcMarks[key] = &apiv1.LisOfIntMarks{Marks: values}
	}

	return &apiv1.DayMarksResponse{Marks: grpcMarks, WorstMark: dayMarks.WorstMark}, nil
}

func (s *serverAPI) GetAverageMarks(ctx context.Context, req *apiv1.AverageMarksRequest) (*apiv1.AverageMarksResponse, error) {
	if err := validateUUID4(req.GetUserToken(), "user token"); err != nil {
		return nil, err
	}
	if err := validateUUID4(req.GetStudentToken(), "student token"); err != nil {
		return nil, err
	}
	if req.GetPeriod() == emptyValue {
		return nil, status.Error(codes.InvalidArgument, "period required")
	}

	avgMarks, err := s.marks.GetAverageMarks(ctx, req.GetUserToken(), req.GetStudentToken(), req.GetPeriod())

	if err != nil {
		if errors.Is(err, service.ErrStudentNotFound) {
			return nil, status.Error(codes.InvalidArgument, "no such student")
		}

		return nil, status.Error(codes.Internal, "failed to get marks")
	}

	return &apiv1.AverageMarksResponse{Marks: avgMarks.Marks, WorstMark: avgMarks.WorstMark}, nil
}

func (s *serverAPI) GetFinalMarks(ctx context.Context, req *apiv1.FinalMarksRequest) (*apiv1.FinalMarksResponse, error) {
	if err := validateUUID4(req.GetUserToken(), "user token"); err != nil {
		return nil, err
	}
	if err := validateUUID4(req.GetStudentToken(), "student token"); err != nil {
		return nil, err
	}

	finalMarks, err := s.marks.GetFinalMarks(ctx, req.GetUserToken(), req.GetStudentToken())

	if err != nil {
		if errors.Is(err, service.ErrStudentNotFound) {
			return nil, status.Error(codes.InvalidArgument, "no such student")
		}

		return nil, status.Error(codes.Internal, "failed to get marks")
	}

	grpcMarks := make(map[string]*apiv1.LisOfIntMarks)

	for key, values := range finalMarks.Marks {
		grpcMarks[key] = &apiv1.LisOfIntMarks{Marks: values}
	}

	return &apiv1.FinalMarksResponse{Marks: grpcMarks, WorstMark: finalMarks.WorstMark}, nil
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
