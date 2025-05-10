package server

import (
	"context"

	pb "your-company.com/project/specs/proto/users"

	"your-company.com/project/services/users/entity"
)

func (s *Server) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {
	loginEntity := entity.MakeLoginPbToEntity(req)

	login, err := s.useCase.Login(ctx, loginEntity)
	if err != nil {
		return nil, err
	}

	return entity.MakeLoginEntityToPb(login), nil
}
