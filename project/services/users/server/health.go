package server

import (
	"context"
	
	"your-company.com/project/services/users/entity"

	pb "your-company.com/project/specs/proto/users"
)

func (s *Server) HealthCheck(ctx context.Context, req *pb.HealthCheckReq) (*pb.HealthCheckResp, error) {
	health, err := s.useCase.HealthCheck(ctx)
	if err != nil {
		return nil, err
	}
	return entity.MakeHealthEntityToPb(health), nil
}
