package handlers

import (
	"google.golang.org/grpc"
)

type Config struct {
	Client *grpc.ClientConn
}
