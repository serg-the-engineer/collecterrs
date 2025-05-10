package server

import (
	"your-company.com/project/pkg/grpcx"

	"your-company.com/project/config/services/otp"

	"google.golang.org/grpc"

	"your-company.com/project/services/otp/usecase"
	pb "your-company.com/project/specs/proto/otp"
)

type Server struct {
	pb.UnimplementedOtpServer

	cfg     *otp.Config
	useCase usecase.Otp
}

func NewServerOptions(useCases usecase.Otp, cfg *otp.Config) *Server {
	return &Server{
		cfg:     cfg,
		useCase: useCases,
	}
}

func (s *Server) NewServer(cfg *grpcx.Config) (*grpc.Server, error) {
	options, err := grpcx.SetOptions(cfg)
	if err != nil {
		return nil, err
	}

	allOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpcx.ProjectErrorInterceptor),
	}
	allOptions = append(allOptions, options...)
	srv := grpc.NewServer(allOptions...)

	pb.RegisterOtpServer(srv, s)

	return srv, nil
}
