package server

import (
	v1 "github.com/toheart/goanalysis/api/analysis/v1"
	"github.com/toheart/goanalysis/internal/conf"
	"github.com/toheart/goanalysis/internal/service"

	nhttp "net/http"
	"os"
	"path/filepath"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, analysis *service.AnalysisService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
		),
		http.Filter(handlers.CORS(
			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
			handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		)),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)

	// 注册API服务
	v1.RegisterAnalysisHTTPServer(srv, analysis)

	// 添加前端静态文件服务
	frontendDir := "./frontweb/dist"
	if _, err := os.Stat(frontendDir); !os.IsNotExist(err) {
		fileServer := nhttp.FileServer(nhttp.Dir(frontendDir))
		srv.HandlePrefix("/", nhttp.HandlerFunc(func(w nhttp.ResponseWriter, r *nhttp.Request) {
			// 检查请求的路径是否存在
			path := filepath.Join(frontendDir, r.URL.Path)
			_, err := os.Stat(path)

			// 如果文件不存在且不是API路径，则返回index.html（处理SPA路由）
			if os.IsNotExist(err) && !isAPIPath(r.URL.Path) {
				nhttp.ServeFile(w, r, filepath.Join(frontendDir, "index.html"))
				return
			}

			fileServer.ServeHTTP(w, r)
		}))
	} else {
		log.NewHelper(logger).Warnf("%s not found", frontendDir)
	}

	return srv
}

// isAPIPath 判断是否为API路径
func isAPIPath(path string) bool {
	// 假设所有API路径都以/api/开头
	return len(path) >= 5 && path[:5] == "/api/"
}
