package server

import (
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, NewServerList)

func NewServerList(http *HttpServer, grpc *GrpcServer) []transport.Server {
	return []transport.Server{
		http,
		grpc,
	}
}
