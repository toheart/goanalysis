package iface

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type InitGrpcHttp interface {
	RegisterHttp(mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error
	RegisterGrpc(svr *grpc.Server)
}
