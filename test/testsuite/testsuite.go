package testsuite

import (
	"AuthService/internal/config"
	"AuthService/internal/pb"
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Suite struct {
	*testing.T
	Cfg        *config.Config
	AuthClient pb.UserServiceClient
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	cfg, err := config.LoadConfig()

	if err != nil {
		t.Fatalf("config load failed: %v", err)
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Hour)

	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	client, err := grpc.DialContext(context.Background(),
		grpcAddress(cfg),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc server connection failed: %v", err)
	}

	return ctx, &Suite{
		T:          t,
		Cfg:        cfg,
		AuthClient: pb.NewUserServiceClient(client),
	}
}

func grpcAddress(cfg *config.Config) string {
	return net.JoinHostPort("localhost", strconv.Itoa(cfg.Port))
}
