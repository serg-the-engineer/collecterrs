package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"your-company.com/project/pkg/redis"

	cfgOtp "your-company.com/project/config/services/otp"
	pbOtp "your-company.com/project/specs/proto/otp"

	"your-company.com/project/pkg/logs"

	"your-company.com/project/pkg/grpcx"

	"your-company.com/project/config/services/users"
	"your-company.com/project/services/users/server"
	"your-company.com/project/services/users/storage"

	"your-company.com/project/services/users/usecase"
)

func main() {
	time.Local = time.UTC
	cfg := users.LoadDefault()

	logger := logs.Logger(cfg.Logger)
	defer logger.Fatal().Msgf("application stopped")

	ctx, cancel := context.WithCancel(logger.WithContext(context.TODO()))

	if err := run(ctx, cancel, cfg); err != nil {
		logger.Info().Err(err).Msg("unable to start application")
	}
}

func run(ctx context.Context, cancel context.CancelFunc, cfg *users.Config) error {
	// Initialize Database
	dbStorage, err := databaseStorage(ctx, cfg)
	if err != nil {
		return fmt.Errorf("unable to init db storage: %w", err)
	}

	// Initialize Redis cache
	redis, err := redis.NewClient(ctx, cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}
	defer redis.Close(ctx)

	// Initialize gRPC clients
	otpServer, err := grpcx.ConnectServer(cfgOtp.Load().GRPC)
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer otpServer.Close()
	otpClient := pbOtp.NewOtpClient(otpServer)

	providers := users.Providers{
		Otp:     otpClient,
		Storage: dbStorage,
		Redis:   redis,
	}

	// Initialize Use cases
	useCases := usecase.New(cfg, &providers)

	// Initialize gRPC server
	service := server.NewServerOptions(useCases, cfg)
	grpcServer, err := service.NewServer(cfg.GRPC)
	if err != nil {
		log.Fatalf("Failed to create gRPC Server: %v", err)
	}

	return grpcx.StartServer(ctx, cfg.GRPC, grpcServer)
}
func databaseStorage(ctx context.Context, cfg *users.Config) (storage.Storage, error) {
	return storage.New(), nil
}
