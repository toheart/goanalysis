package session

import (
	"context"
	"net/http"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/metadata"
)

const (
	// HTTP Header 前缀
	AnalysisHeaderPrefix = "x-analysis-"

	// 具体的键名（HTTP header 和 gRPC metadata 使用相同的键名）
	AnalysisDBPathKey    = "x-analysis-db-path"
	AnalysisSessionIDKey = "x-analysis-session-id"
)

// GRPCGatewayMiddleware 为grpc-gateway创建中间件，将session信息注入到gRPC metadata中
func GRPCGatewayMiddleware(manager *Manager, logger log.Logger) func(http.Handler) http.Handler {
	logHelper := log.NewHelper(logger)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 确保session有ID，但不立即保存
			sessionID, needsSave, err := manager.EnsureSessionIDWithoutSave(r)
			if err != nil {
				logHelper.Errorf("Failed to ensure session ID: %v", err)
				// 如果session创建失败，返回HTTP错误，不要继续处理
				http.Error(w, "Session initialization failed", http.StatusInternalServerError)
				return
			}

			// 从session中获取数据库路径
			dbPath, exists, err := manager.GetAnalysisPath(r)
			if err != nil {
				logHelper.Warnf("Failed to get analysis path from session %s: %v", sessionID, err)
				// 这里不返回错误，因为可能是首次访问，还没有设置分析路径
			}

			// 如果存在数据库路径，将其添加到请求头中，这样grpc-gateway会将其转换为gRPC metadata
			if exists && dbPath != "" {
				r.Header.Set(AnalysisDBPathKey, dbPath)
				logHelper.Infof("Session %s: Set header %s=%s for URL: %s", sessionID, AnalysisDBPathKey, dbPath, r.URL.Path)
			} else {
				logHelper.Debugf("Session %s: No DB path found (exists: %v, dbPath: '%s') for URL: %s", sessionID, exists, dbPath, r.URL.Path)
			}

			// 将session ID添加到请求头中
			r.Header.Set(AnalysisSessionIDKey, sessionID)

			// 处理请求
			next.ServeHTTP(w, r)

			// 请求处理完成后，检查是否需要保存session
			if needsSave || manager.SessionNeedsSave(r) {
				if err := manager.SaveSession(w, r); err != nil {
					logHelper.Errorf("Failed to save session: %v", err)
				}
			}
		})
	}
}

// GetMetadataFromGRPCContext 通用的从 gRPC context 中获取 metadata 的函数
func GetMetadataFromGRPCContext(ctx context.Context, key string) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}

	values := md.Get(key)
	if len(values) == 0 {
		return "", false
	}

	return values[0], true
}

// GetDBPathFromGRPCContext 从gRPC context中获取数据库路径
func GetDBPathFromGRPCContext(ctx context.Context) (string, bool) {
	return GetMetadataFromGRPCContext(ctx, AnalysisDBPathKey)
}

// GetSessionIDFromGRPCContext 从gRPC context中获取session ID
func GetSessionIDFromGRPCContext(ctx context.Context) string {
	value, _ := GetMetadataFromGRPCContext(ctx, AnalysisSessionIDKey)
	return value
}
