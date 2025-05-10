package grpcx

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

const maxCallRecvMsgSize = 1024 * 1024 * 20

func ConnectServer(cfg *Config) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxCallRecvMsgSize),
			grpc.MaxCallSendMsgSize(maxCallRecvMsgSize),
		),
	}

	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	keepaliveParams := keepalive.ClientParameters{
		Time:                cfg.KeepAliveTime,
		Timeout:             cfg.KeepAliveTimeout,
		PermitWithoutStream: true,
	}
	opts = append(opts, grpc.WithKeepaliveParams(keepaliveParams))

	ctx := context.Background()

	// Подключение к gRPC серверу
	conn, err := grpc.DialContext(ctx, cfg.Addr(), opts...)
	if err != nil {
		log.Error().Msgf("failed to connect to GRPC server at address: %s, error: %v", cfg.Addr(), err)

		return nil, fmt.Errorf("failed to connect to GRPC server at address %s: %w", cfg.Addr(), err)
	}

	log.Info().Msgf("attempting to connect to GRPC server at address: %s", cfg.Addr())

	return conn, nil
}
