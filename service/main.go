package main

import (
	"game_server/pb"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	grpcServer := grpc.NewServer()
	pb.RegisterGameServiceServer(grpcServer, new(GameServiceServer))

	lis, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal(err)
	}
	grpcServer.Serve(lis)
}
