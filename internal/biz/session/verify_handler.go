package session

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	v1 "github.com/toheart/goanalysis/api/analysis/v1"
	staticv1 "github.com/toheart/goanalysis/api/staticanalysis/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

// RegisterVerifyPathHandler 注册自定义的验证路径处理器
// 这个处理器会在验证成功时自动将路径保存到session中
func RegisterVerifyPathHandler(mux *runtime.ServeMux, manager *Manager, endpoint string, opts []grpc.DialOption, logger log.Logger) error {
	logHelper := log.NewHelper(logger)

	return mux.HandlePath("POST", "/api/runtime/verify/path", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		// 读取请求体
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logHelper.Errorf("Failed to read request body: %v", err)
			http.Error(w, "Failed to read request", http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))

		// 解析请求以获取路径
		var req v1.VerifyProjectPathReq
		if err := protojson.Unmarshal(body, &req); err != nil {
			logHelper.Errorf("Failed to parse request: %v", err)
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		// 创建gRPC连接并调用原始的验证方法
		conn, err := grpc.Dial(endpoint, opts...)
		if err != nil {
			logHelper.Errorf("Failed to connect to gRPC server: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		client := v1.NewAnalysisClient(conn)

		// 调用gRPC方法进行验证
		resp, err := client.VerifyProjectPath(context.Background(), &req)
		if err != nil {
			logHelper.Errorf("Failed to verify path: %v", err)
			http.Error(w, "Verification failed", http.StatusInternalServerError)
			return
		}

		// 如果验证成功，将路径保存到session中
		if resp.Verified {
			if err := manager.SetAnalysisPath(w, r, req.Path); err != nil {
				logHelper.Errorf("Failed to set path in session: %v", err)
			} else {
				logHelper.Infof("Path verified and cached to session: %s", req.Path)
			}
		}

		// 返回验证结果
		w.Header().Set("Content-Type", "application/json")
		respBytes, err := protojson.Marshal(resp)
		if err != nil {
			logHelper.Errorf("Failed to marshal response: %v", err)
			http.Error(w, "Failed to create response", http.StatusInternalServerError)
			return
		}

		w.Write(respBytes)
	})
}

// RegisterAnalyzeDbFileHandler 注册自定义的数据库分析处理器
// 这个处理器会在分析成功时自动将数据库路径保存到session中
func RegisterAnalyzeDbFileHandler(mux *runtime.ServeMux, manager *Manager, endpoint string, opts []grpc.DialOption, logger log.Logger) error {
	logHelper := log.NewHelper(logger)

	return mux.HandlePath("POST", "/api/static/analyze", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		// 读取请求体
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logHelper.Errorf("Failed to read request body: %v", err)
			http.Error(w, "Failed to read request", http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))

		// 解析请求以获取数据库路径
		var req staticv1.AnalyzeDbFileRequest
		if err := protojson.Unmarshal(body, &req); err != nil {
			logHelper.Errorf("Failed to parse request: %v", err)
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		// 创建gRPC连接并调用原始的分析方法
		conn, err := grpc.Dial(endpoint, opts...)
		if err != nil {
			logHelper.Errorf("Failed to connect to gRPC server: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		client := staticv1.NewStaticAnalysisClient(conn)

		// 调用gRPC方法进行分析
		resp, err := client.AnalyzeDbFile(context.Background(), &req)
		if err != nil {
			logHelper.Errorf("Failed to analyze database: %v", err)
			http.Error(w, "Analysis failed", http.StatusInternalServerError)
			return
		}

		// 分析成功后，将数据库路径保存到session中
		if err := manager.SetAnalysisPath(w, r, req.DbPath); err != nil {
			logHelper.Errorf("Failed to set db path in session: %v", err)
		} else {
			logHelper.Infof("Database analyzed and path cached to session: %s", req.DbPath)
		}

		// 返回分析结果
		w.Header().Set("Content-Type", "application/json")
		respBytes, err := protojson.Marshal(resp)
		if err != nil {
			logHelper.Errorf("Failed to marshal response: %v", err)
			http.Error(w, "Failed to create response", http.StatusInternalServerError)
			return
		}

		w.Write(respBytes)
	})
}
