package suite

import (
	"Elschool-API/internal/config"
	"context"
	apiv1 "github.com/Ilya-Repin/elschooler/protos/gen/api/api.v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"strconv"
	"testing"
)

const (
	grpcHost = "localhost"
)

type Suite struct {
	*testing.T
	Cfg           *config.Config
	UserClient    apiv1.UserClient
	StudentClient apiv1.StudentClient
	MarksClient   apiv1.MarksClient
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	cfg := config.MustLoadByPath("../config/local_tests.yaml")

	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPCConfig.Timeout)

	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	cc, err := grpc.DialContext(ctx,
		grpcAddress(cfg),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc server connection failed: %v", err)
	}

	return ctx, &Suite{
		T:             t,
		Cfg:           cfg,
		UserClient:    apiv1.NewUserClient(cc),
		StudentClient: apiv1.NewStudentClient(cc),
		MarksClient:   apiv1.NewMarksClient(cc),
	}
}

func grpcAddress(cfg *config.Config) string {
	return net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPCConfig.Port))
}
