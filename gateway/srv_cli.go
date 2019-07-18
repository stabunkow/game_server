package main

import (
	"game_server/pb"
	"log"
	"sync"

	"google.golang.org/grpc"
)

var defaultGameServiceClient pb.GameServiceClient
var defaultGameServiceClientOnce sync.Once

func GetGameServiceClient() pb.GameServiceClient {
	defaultGameServiceClientOnce.Do(func() {
		conn, err := grpc.Dial(":1234", grpc.WithInsecure())

		if err != nil {
			log.Println(err)
		}

		defaultGameServiceClient = pb.NewGameServiceClient(conn)
	})

	return defaultGameServiceClient
}
