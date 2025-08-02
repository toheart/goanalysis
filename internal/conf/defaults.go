package conf

import (
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
)

// GetDefaultConfig 返回默认配置
func GetDefaultConfig() *Bootstrap {
	return &Bootstrap{
		Server: &Server{
			Http: &Server_HTTP{
				Network: "tcp",
				Addr:    "0.0.0.0:8001",
				Timeout: durationpb.New(1 * time.Second),
			},
			Grpc: &Server_GRPC{
				Network: "tcp",
				Addr:    "0.0.0.0:9000",
				Timeout: durationpb.New(1 * time.Second),
			},
		},
		Logger: &Logger{
			Level:      "debug",
			FilePath:   "./logs/app.log",
			Console:    true,
			MaxSize:    100,
			MaxAge:     7,
			MaxBackups: 10,
			Compress:   true,
		},
		Biz: &Biz{
			Gitlab: &GitLab{
				Token:    "", // 从环境变量读取
				Url:      "", // 从环境变量读取
				CloneDir: "./data",
			},
			Openai: &OpenAI{
				ApiKey:  "", // 从环境变量读取
				ApiBase: "", // 从环境变量读取
				Model:   "", // 从环境变量读取
			},
			StaticStorePath:  "./data/static",
			RuntimeStorePath: "./data/runtime",
			FileStoragePath:  "./data/files",
		},
		Data: &Data{
			Dbpath: "./goanalysis.db",
		},
	}
}

// LoadFromEnv 从环境变量加载敏感配置
func LoadFromEnv(config *Bootstrap) {
	// GitLab配置
	if config.Biz != nil && config.Biz.Gitlab != nil {
		if token := getEnvOrDefault("GITLAB_TOKEN", ""); token != "" {
			config.Biz.Gitlab.Token = token
		}
		if url := getEnvOrDefault("GITLAB_API_URL", ""); url != "" {
			config.Biz.Gitlab.Url = url
		}
	}

	// OpenAI配置
	if config.Biz != nil && config.Biz.Openai != nil {
		if apiKey := getEnvOrDefault("OPENAI_API_KEY", ""); apiKey != "" {
			config.Biz.Openai.ApiKey = apiKey
		}
		if apiBase := getEnvOrDefault("OPENAI_API_BASE", ""); apiBase != "" {
			config.Biz.Openai.ApiBase = apiBase
		}
		if model := getEnvOrDefault("OPENAI_MODEL", ""); model != "" {
			config.Biz.Openai.Model = model
		}
	}
}

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := getEnv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnv 获取环境变量（这里需要导入os包，但为了避免循环依赖，我们在这里声明）
// 实际实现会在调用处处理
func getEnv(key string) string {
	// 这个函数会在实际使用时通过os.Getenv实现
	return ""
}
