package server

import (
	"context"
	"net/http"

	"github.com/go-kratos/kratos/v2/transport"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/toheart/goanalysis/internal/conf"
	"github.com/toheart/goanalysis/internal/server/iface"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/rs/cors"
)

var _ transport.Server = (*HttpServer)(nil)

type HttpServer struct {
	server *http.Server
	log    *log.Helper
}

func (h *HttpServer) Start(ctx context.Context) error {
	h.log.Infof("start http gateway server: %s", h.server.Addr)
	return h.server.ListenAndServe()
}

func (h *HttpServer) Stop(ctx context.Context) error {
	h.log.Infof("Shutting down the http gateway server")
	if err := h.server.Shutdown(ctx); err != nil {
		h.log.Errorf("Failed to shutdown http gateway server: %v", err)
	}
	return nil
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, logger log.Logger, services ...iface.InitGrpcHttp) *HttpServer {
	h := &HttpServer{
		log: log.NewHelper(log.With(logger, "module", "http")),
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	mux := runtime.NewServeMux()

	// 创建一个支持CORS的处理器
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // 允许所有来源，生产环境中应该限制为特定域名
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           86400, // 预检请求结果缓存24小时
	}).Handler(mux)

	h.server = &http.Server{
		Addr:    c.Http.Addr,
		Handler: corsHandler, // 使用CORS处理器包装mux
	}
	for _, item := range services {
		if err := item.RegisterHttp(mux, c.Grpc.Addr, opts); err != nil {
			panic(err)
		}
	}
	// 添加Prometheus 接口
	err := mux.HandlePath("GET", "/metrics", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		promhttp.Handler().ServeHTTP(w, r)
	})
	if err != nil {
		panic(err)
	}
	return h
}

// isAPIPath 判断是否为API路径
func isAPIPath(path string) bool {
	// 假设所有API路径都以/api/开头
	return len(path) >= 5 && path[:5] == "/api/"
}
