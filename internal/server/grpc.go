package server

import (
	"context"
	"fmt"
	"net"
	"runtime/debug"

	"github.com/go-kratos/kratos/v2/transport"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/toheart/goanalysis/internal/conf"
	"github.com/toheart/goanalysis/internal/server/iface"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"

	"github.com/go-kratos/kratos/v2/log"
)

var _ transport.Server = (*GrpcServer)(nil)

type GrpcServer struct {
	conf *conf.Server
	log  *log.Helper

	svr *grpc.Server
}

func (g *GrpcServer) Start(ctx context.Context) error {
	l, err := net.Listen("tcp", g.conf.Grpc.Addr)
	if err != nil {
		return err
	}
	g.log.Infow("msg", "starting gRPC server", "addr", l.Addr().String())
	return g.svr.Serve(l)
}

func (g *GrpcServer) Stop(ctx context.Context) error {
	g.svr.GracefulStop()
	g.svr.Stop()
	return nil
}

// interceptorLogger adapts go-kit logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func interceptorLogger(l *log.Helper) logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		largs := append([]any{"msg", msg}, fields...)
		switch lvl {
		case logging.LevelDebug:
			l.Debugw(largs...)
		case logging.LevelInfo:
			l.Infow(largs)
		case logging.LevelWarn:
			l.Warnw(largs)
		case logging.LevelError:
			l.Errorw(largs)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, logger log.Logger, services ...iface.InitGrpcHttp) *GrpcServer {
	g := &GrpcServer{
		conf: c,
		log:  log.NewHelper(log.With(logger, "module", "grpc")),
	}
	var opts []grpc.ServerOption
	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)
	// Setup metric for panic recoveries.
	panicsTotal := promauto.NewCounter(prometheus.CounterOpts{
		Name: "grpc_req_panics_recovered_total",
		Help: "Total number of gRPC requests recovered from internal panic.",
	})
	grpcPanicRecoveryHandler := func(p any) (err error) {
		panicsTotal.Inc()
		g.log.Errorw("msg", "recovered from panic", "panic", p, "stack", debug.Stack())
		return status.Errorf(codes.Internal, "%s", p)
	}
	prometheus.MustRegister(srvMetrics)
	// 设置traceId
	logTraceID := func(ctx context.Context) logging.Fields {
		if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
			return logging.Fields{"traceID", span.TraceID().String()}
		}
		return nil
	}
	// 设置拦截器
	opts = append(opts, grpc.ChainUnaryInterceptor(
		srvMetrics.UnaryServerInterceptor(),
		logging.UnaryServerInterceptor(interceptorLogger(g.log), logging.WithFieldsFromContext(logTraceID)),
		recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
	))
	opts = append(opts,
		grpc.StatsHandler(otelgrpc.NewServerHandler()))

	g.svr = grpc.NewServer(opts...)
	for _, item := range services {
		item.RegisterGrpc(g.svr)
	}
	t := &testpb.TestPingService{}
	testpb.RegisterTestServiceServer(g.svr, t)
	return g
}
