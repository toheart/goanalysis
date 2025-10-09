package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-kratos/kratos/v2/transport"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rakyll/statik/fs"
	"github.com/toheart/goanalysis/internal/biz/entity"
	"github.com/toheart/goanalysis/internal/biz/filemanager"
	"github.com/toheart/goanalysis/internal/biz/session"
	"github.com/toheart/goanalysis/internal/biz/staticanalysis"
	"github.com/toheart/goanalysis/internal/conf"
	"github.com/toheart/goanalysis/internal/server/iface"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/rs/cors"

	// 导入生成的 statik 包
	_ "github.com/toheart/goanalysis/statik"
)

var _ transport.Server = (*HttpServer)(nil)

type HttpServer struct {
	server     *http.Server
	log        *log.Helper
	staticBiz  *staticanalysis.StaticAnalysisBiz
	fileBiz    *filemanager.FileBiz
	sessionMgr *session.Manager
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

// handleAnalysisEvents 处理分析事件流
func (h *HttpServer) handleAnalysisEvents(w http.ResponseWriter, r *http.Request) {
	// 从URL路径中获取taskId
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}
	taskId := parts[len(parts)-1]

	h.log.Infof("Starting event stream for task: %s", taskId)

	// 设置 SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// 获取任务状态通道
	statusChan, err := h.staticBiz.GetTaskStatusChan(taskId)
	if err != nil {
		h.log.Errorf("Failed to get status channel for task %s: %v", taskId, err)
		http.Error(w, "Failed to get status channel", http.StatusInternalServerError)
		return
	}

	// 创建一个done通道用于处理客户端断开连接
	done := make(chan bool)
	notify := r.Context().Done()
	go func() {
		<-notify
		h.log.Infof("Client disconnected from event stream for task: %s", taskId)
		done <- true
	}()

	// 发送初始连接消息
	initialMsg := entity.AnalysisEvent{
		Type:    entity.TaskStatusStarting,
		Message: "Analysis task started",
	}
	if err := sendSSEEvent(w, initialMsg); err != nil {
		h.log.Errorf("Failed to send initial message: %v", err)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// 标记是否已经发送了完成消息
	completedSent := false

	// 监听消息和完成信号
	for {
		select {
		case msg, ok := <-statusChan:
			// 如果通道已关闭，发送完成消息并退出
			if !ok {
				if !completedSent {
					completedMsg := entity.AnalysisEvent{
						Type:    entity.TaskStatusCompleted,
						Message: "Analysis task completed",
					}
					if err := sendSSEEvent(w, completedMsg); err != nil {
						h.log.Errorf("Failed to send completion message: %v", err)
					}
					flusher.Flush()
					completedSent = true
				}
				h.log.Infof("Status channel closed for task: %s", taskId)
				return
			}

			// 发送正常消息
			data := entity.AnalysisEvent{
				Type:    entity.TaskStatusProcessing,
				Message: string(msg),
			}
			if err := sendSSEEvent(w, data); err != nil {
				h.log.Errorf("Failed to send message: %v", err)
				return
			}
			flusher.Flush()

		case <-done:
			// 客户端断开连接
			h.log.Infof("Client disconnected, stopping event stream for task: %s", taskId)
			return
		}
	}
}

// sendSSEEvent 发送SSE事件
func sendSSEEvent(w http.ResponseWriter, data interface{}) error {
	// 将数据转换为JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// 写入SSE格式的数据
	_, err = fmt.Fprintf(w, "data: %s\n\n", jsonData)
	return err
}

// createSPAFileServer 创建SPA友好的文件服务器
func (h *HttpServer) createSPAFileServer(statikFS http.FileSystem) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 尝试打开请求的文件
		file, err := statikFS.Open(r.URL.Path)
		if err == nil {
			// 文件存在，提供文件内容
			defer file.Close()

			// 获取文件信息
			stat, err := file.Stat()
			if err == nil && !stat.IsDir() {
				http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
				return
			}
		}

		// 文件不存在或是目录，对于SPA应用返回index.html
		indexFile, err := statikFS.Open("/index.html")
		if err != nil {
			// 如果连index.html都不存在，返回404
			http.NotFound(w, r)
			return
		}
		defer indexFile.Close()

		stat, err := indexFile.Stat()
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// 设置正确的Content-Type
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, stat.Name(), stat.ModTime(), indexFile)
	})
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, logger log.Logger, staticBiz *staticanalysis.StaticAnalysisBiz, fileBiz *filemanager.FileBiz, sessionMgr *session.Manager, services ...iface.InitGrpcHttp) *HttpServer {
	logHelper := log.NewHelper(log.With(logger, "module", "http"))

	h := &HttpServer{
		log:        logHelper,
		staticBiz:  staticBiz,
		fileBiz:    fileBiz,
		sessionMgr: sessionMgr,
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// 创建通用的 header 匹配器，允许所有 x-analysis- 前缀的 header 传递到 gRPC metadata
	sessionHeaderMatcher := func(key string) (string, bool) {
		lowerKey := strings.ToLower(key)
		if strings.HasPrefix(lowerKey, session.AnalysisHeaderPrefix) {
			// 直接返回小写的键名，保持一致
			return lowerKey, true
		}
		return "", false
	}

	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(sessionHeaderMatcher),
	)

	// 创建一个支持CORS的处理器
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8081", "http://192.168.141.128:8082"}, // 允许前端域名和后端IP访问
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Set-Cookie"}, // 暴露Set-Cookie头给前端
		AllowCredentials: true,
		MaxAge:           86400, // 预检请求结果缓存24小时
	})

	// 创建自定义的 HTTP 处理器
	handler := http.NewServeMux()

	// 添加 SSE 端点
	handler.HandleFunc("/api/static/analysis/", h.handleAnalysisEvents)
	logHelper.Infof("SSE endpoint registered: /api/static/analysis/{taskId}")

	// Session相关的API通过gRPC+grpc-gateway提供，不需要单独的HTTP处理器

	// 使用 statik 嵌入的静态文件
	statikFS, err := fs.New()
	if err != nil {
		logHelper.Fatalf("Failed to create statik filesystem: %v", err)
	}

	// 创建SPA友好的文件服务器
	fileServer := h.createSPAFileServer(statikFS)

	// 创建一个处理所有请求的处理器
	rootHandler := h.BaseHandler(mux, fileServer)

	// 将根处理器包装在session中间件和CORS处理器中
	// 使用grpc-gateway中间件来处理session信息注入到gRPC metadata
	grpcGatewayMiddleware := session.GRPCGatewayMiddleware(sessionMgr, logger)
	handler.Handle("/", corsHandler.Handler(grpcGatewayMiddleware(rootHandler)))
	logHelper.Infof("Root handler with CORS and embedded static file serving registered")

	serverAddr := c.Http.Addr
	h.server = &http.Server{
		Addr:    serverAddr,
		Handler: handler,
	}

	logHelper.Infof("HTTP server configuration completed, listening address: %s", serverAddr)

	for _, item := range services {
		if err := item.RegisterHttp(mux, "127.0.0.1:9000", opts); err != nil {
			panic(err)
		}
	}

	// 注册自定义的验证路径处理器（验证成功时自动缓存到session）
	if err := session.RegisterVerifyPathHandler(mux, sessionMgr, "127.0.0.1:9000", opts, logger); err != nil {
		logHelper.Errorf("Failed to register verify path handler: %v", err)
	}

	// 注册自定义的数据库分析处理器（分析成功时自动缓存数据库路径到session）
	if err := session.RegisterAnalyzeDbFileHandler(mux, sessionMgr, "127.0.0.1:9000", opts, logger); err != nil {
		logHelper.Errorf("Failed to register analyze db file handler: %v", err)
	}

	// 添加Prometheus 接口
	err = mux.HandlePath("GET", "/metrics", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		promhttp.Handler().ServeHTTP(w, r)
	})
	if err != nil {
		panic(err)
	}

	return h
}
