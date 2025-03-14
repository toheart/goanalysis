package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-kratos/kratos/v2/transport"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/toheart/goanalysis/internal/biz/entity"
	"github.com/toheart/goanalysis/internal/biz/filemanager"
	"github.com/toheart/goanalysis/internal/biz/staticanalysis"
	"github.com/toheart/goanalysis/internal/conf"
	"github.com/toheart/goanalysis/internal/server/iface"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/rs/cors"
)

var _ transport.Server = (*HttpServer)(nil)

type HttpServer struct {
	server    *http.Server
	log       *log.Helper
	staticBiz *staticanalysis.StaticAnalysisBiz
	fileBiz   *filemanager.FileBiz
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

// 添加一个辅助函数来检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, logger log.Logger, staticBiz *staticanalysis.StaticAnalysisBiz, services ...iface.InitGrpcHttp) *HttpServer {
	logHelper := log.NewHelper(log.With(logger, "module", "http"))

	h := &HttpServer{
		log:       logHelper,
		staticBiz: staticBiz,
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
	})

	// 创建自定义的 HTTP 处理器
	handler := http.NewServeMux()

	// 添加 SSE 端点
	handler.HandleFunc("/api/static/analysis/", h.handleAnalysisEvents)
	logHelper.Infof("SSE endpoint registered: /api/static/analysis/{taskId}")

	// 定义前端目录
	frontendDir := "./web"
	fileServer := http.FileServer(http.Dir(frontendDir))

	// 创建一个处理所有请求的处理器
	rootHandler := h.BaseHandler(mux, frontendDir, fileServer)

	// 将根处理器包装在CORS处理器中
	handler.Handle("/", corsHandler.Handler(rootHandler))
	logHelper.Infof("Root handler with CORS and static file serving registered")

	serverAddr := c.Http.Addr
	h.server = &http.Server{
		Addr:    serverAddr,
		Handler: handler,
	}

	logHelper.Infof("HTTP server configuration completed, listening address: %s", serverAddr)

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
