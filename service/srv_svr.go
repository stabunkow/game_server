package main

import (
	"context"
	"errors"
	"game_server/model"
	"game_server/pb"
)

type GameServiceServer struct{}

func (s *GameServiceServer) GetUserInfo(ctx context.Context, arg *pb.String) (*pb.User, error) {
	uid := arg.GetValue()
	usr := model.GetUserById(uid)

	if usr == nil {
		return nil, errors.New("user not found")
	}

	return &pb.User{
		Id:        usr.GetId(),
		Email:     usr.GetEmail(),
		CreatedAt: usr.GetCreatedAt(),
		UpdatedAt: usr.GetUpdatedAt(),
	}, nil
}
