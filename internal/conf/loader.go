package conf

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/durationpb"
)

// LoadConfig 加载配置，优先级：命令行参数 > 环境变量 > 配置文件 > 默认值
func LoadConfig(cmd *cobra.Command) (*Bootstrap, error) {
	// 1. 加载默认配置
	config := GetDefaultConfig()

	// 2. 如果指定了配置文件，加载配置文件
	if configFile := cmd.Flag("conf").Value.String(); configFile != "" {
		if err := loadFromFile(configFile, config); err != nil {
			return nil, err
		}
	}

	// 3. 从环境变量加载敏感信息
	loadFromEnv(config)

	// 4. 命令行参数覆盖（最高优先级）
	loadFromFlags(cmd, config)

	return config, nil
}

// loadFromFile 从配置文件加载配置
func loadFromFile(configFile string, bootstrap *Bootstrap) error {
	c := config.New(
		config.WithSource(
			file.NewSource(configFile),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		return err
	}

	return c.Scan(bootstrap)
}

// loadFromEnv 从环境变量加载配置
func loadFromEnv(config *Bootstrap) {
	// GitLab配置
	if config.Biz != nil && config.Biz.Gitlab != nil {
		if token := os.Getenv("GITLAB_TOKEN"); token != "" {
			config.Biz.Gitlab.Token = token
		}
		if url := os.Getenv("GITLAB_API_URL"); url != "" {
			config.Biz.Gitlab.Url = url
		}
	}

	// OpenAI配置
	if config.Biz != nil && config.Biz.Openai != nil {
		if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
			config.Biz.Openai.ApiKey = apiKey
		}
		if apiBase := os.Getenv("OPENAI_API_BASE"); apiBase != "" {
			config.Biz.Openai.ApiBase = apiBase
		}
		if model := os.Getenv("OPENAI_MODEL"); model != "" {
			config.Biz.Openai.Model = model
		}
	}
}

// loadFromFlags 从命令行参数加载配置
func loadFromFlags(cmd *cobra.Command, config *Bootstrap) {
	// 服务器配置
	if flag := cmd.Flag("http-addr"); flag != nil && flag.Changed {
		config.Server.Http.Addr = flag.Value.String()
	}
	if flag := cmd.Flag("grpc-addr"); flag != nil && flag.Changed {
		config.Server.Grpc.Addr = flag.Value.String()
	}
	if flag := cmd.Flag("http-timeout"); flag != nil && flag.Changed {
		if timeout, err := time.ParseDuration(flag.Value.String()); err == nil {
			config.Server.Http.Timeout = durationpb.New(timeout)
		}
	}
	if flag := cmd.Flag("grpc-timeout"); flag != nil && flag.Changed {
		if timeout, err := time.ParseDuration(flag.Value.String()); err == nil {
			config.Server.Grpc.Timeout = durationpb.New(timeout)
		}
	}

	// 日志配置
	if flag := cmd.Flag("log-level"); flag != nil && flag.Changed {
		config.Logger.Level = flag.Value.String()
	}
	if flag := cmd.Flag("log-file"); flag != nil && flag.Changed {
		config.Logger.FilePath = flag.Value.String()
	}
	if flag := cmd.Flag("log-console"); flag != nil && flag.Changed {
		if console, err := strconv.ParseBool(flag.Value.String()); err == nil {
			config.Logger.Console = console
		}
	}
	if flag := cmd.Flag("log-max-size"); flag != nil && flag.Changed {
		if maxSize, err := strconv.ParseInt(flag.Value.String(), 10, 32); err == nil {
			config.Logger.MaxSize = int32(maxSize)
		}
	}
	if flag := cmd.Flag("log-max-age"); flag != nil && flag.Changed {
		if maxAge, err := strconv.ParseInt(flag.Value.String(), 10, 32); err == nil {
			config.Logger.MaxAge = int32(maxAge)
		}
	}
	if flag := cmd.Flag("log-max-backups"); flag != nil && flag.Changed {
		if maxBackups, err := strconv.ParseInt(flag.Value.String(), 10, 32); err == nil {
			config.Logger.MaxBackups = int32(maxBackups)
		}
	}
	if flag := cmd.Flag("log-compress"); flag != nil && flag.Changed {
		if compress, err := strconv.ParseBool(flag.Value.String()); err == nil {
			config.Logger.Compress = compress
		}
	}

	// 业务配置
	if flag := cmd.Flag("gitlab-token"); flag != nil && flag.Changed {
		config.Biz.Gitlab.Token = flag.Value.String()
	}
	if flag := cmd.Flag("gitlab-url"); flag != nil && flag.Changed {
		config.Biz.Gitlab.Url = flag.Value.String()
	}
	if flag := cmd.Flag("gitlab-clone-dir"); flag != nil && flag.Changed {
		config.Biz.Gitlab.CloneDir = flag.Value.String()
	}
	if flag := cmd.Flag("openai-api-key"); flag != nil && flag.Changed {
		config.Biz.Openai.ApiKey = flag.Value.String()
	}
	if flag := cmd.Flag("openai-api-base"); flag != nil && flag.Changed {
		config.Biz.Openai.ApiBase = flag.Value.String()
	}
	if flag := cmd.Flag("openai-model"); flag != nil && flag.Changed {
		config.Biz.Openai.Model = flag.Value.String()
	}
	if flag := cmd.Flag("static-store-path"); flag != nil && flag.Changed {
		config.Biz.StaticStorePath = flag.Value.String()
	}
	if flag := cmd.Flag("runtime-store-path"); flag != nil && flag.Changed {
		config.Biz.RuntimeStorePath = flag.Value.String()
	}
	if flag := cmd.Flag("file-storage-path"); flag != nil && flag.Changed {
		config.Biz.FileStoragePath = flag.Value.String()
	}

	// 数据配置
	if flag := cmd.Flag("db-path"); flag != nil && flag.Changed {
		config.Data.Dbpath = flag.Value.String()
	}
}

// ValidateConfig 验证配置
func ValidateConfig(config *Bootstrap) error {
	// 验证服务器配置
	if config.Server == nil {
		config.Server = &Server{}
	}
	if config.Server.Http == nil {
		config.Server.Http = &Server_HTTP{}
	}
	if config.Server.Grpc == nil {
		config.Server.Grpc = &Server_GRPC{}
	}

	// 验证日志配置
	if config.Logger == nil {
		config.Logger = &Logger{}
	}

	// 验证业务配置
	if config.Biz == nil {
		config.Biz = &Biz{}
	}
	if config.Biz.Gitlab == nil {
		config.Biz.Gitlab = &GitLab{}
	}
	if config.Biz.Openai == nil {
		config.Biz.Openai = &OpenAI{}
	}

	// 验证数据配置
	if config.Data == nil {
		config.Data = &Data{}
	}

	return nil
}

// ParseDuration 解析时间字符串
func ParseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	return time.ParseDuration(s)
}
