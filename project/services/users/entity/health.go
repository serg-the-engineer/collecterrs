package entity

import (
	pb "your-company.com/project/specs/proto/users"
)

type (
	Health struct {
		Status bool
	}
)

func MakeHealth(status bool) *Health {
	return &Health{
		Status: status,
	}
}

func MakeHealthEntityToPb(req *Health) *pb.HealthCheckResp {
	return &pb.HealthCheckResp{
		Status: req.Status,
	}
}
