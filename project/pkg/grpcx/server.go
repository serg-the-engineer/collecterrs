package grpcx

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"your-company.com/project/pkg/logs"
)

func StartServer(ctx context.Context, cfg *Config, server *grpc.Server) error {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	logger := logs.FromContext(ctx)
	logger.Info().Msg("starting grpc server")

	if cfg.Reflection {
		logger.Info().Msg("grpc reflection enabled")
		reflection.Register(server)
	}

	serverErrors := make(chan error, 1)

	grpcListener, err := net.Listen("tcp", cfg.AddrServer())
	if err != nil {
		logger.Error().Msg("failed to listen on grpc port")

		return fmt.Errorf("failed to listen on grpc port: %w", err)
	}

	go func() {
		serverErrors <- server.Serve(grpcListener)
	}()
	logger.Info().Msgf("grpc service started at address: %s", cfg.AddrServer())

	select {
	case err := <-serverErrors:
		logger.Info().Msg("grpc server has closed")

		return fmt.Errorf("grpc server has closed: %w", err)
	case sig := <-shutdown:
		logger.Info().Str("signal", sig.String()).Msg("Start shutdown")
		server.GracefulStop()
	case <-ctx.Done():
		logger.Info().Msg("closing grpc server due to context cancellation")
		server.GracefulStop()

		return nil
	}

	return nil
}

func SetOptions(cfg *Config) ([]grpc.ServerOption, error) {
	var serverOptions []grpc.ServerOption

	// Устанавливаем ограничения на размер сообщения
	serverOptions = append(
		serverOptions,
		grpc.MaxRecvMsgSize(cfg.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(cfg.MaxSendMsgSize),
	)

	// Настраиваем параметры keepalive
	serverOptions = append(
		serverOptions,
		grpc.KeepaliveParams(
			keepalive.ServerParameters{
				MaxConnectionIdle:     cfg.MaxConnectionIdle,
				MaxConnectionAge:      cfg.MaxConnectionAge,
				MaxConnectionAgeGrace: cfg.MaxConnectionAgeGrace,
				Time:                  cfg.KeepAliveTime,
				Timeout:               cfg.KeepAliveTimeout,
			},
		),
	)
	return serverOptions, nil
}
