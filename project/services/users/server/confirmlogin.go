package server

import (
	"context"

	pb "your-company.com/project/specs/proto/users"

	"your-company.com/project/services/users/entity"
)

func (s *Server) ConfirmLogin(ctx context.Context, req *pb.ConfirmLoginReq) (*pb.ConfirmLoginResp, error) {
	confirmLoginEntity := entity.MakeConfirmLoginPbToEntity(req)

	confirmLogin, err := s.useCase.ConfirmLogin(ctx, confirmLoginEntity)
	if err != nil {
		return nil, err
	}

	return entity.MakeConfirmLoginEntityToPb(confirmLogin), nil
}
