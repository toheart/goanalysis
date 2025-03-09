package logger

import (
	"os"
	"path/filepath"

	kratoszap "github.com/go-kratos/kratos/contrib/log/zap/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/toheart/goanalysis/internal/conf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger 创建一个新的日志记录器
func NewLogger(c *conf.Logger) log.Logger {
	// 确保日志目录存在
	if c.FilePath != "" {
		logDir := filepath.Dir(c.FilePath)
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			log.DefaultLogger.Log(log.LevelError, "msg", "failed to create log directory", "error", err)
		}
	}

	// 设置日志级别
	var level zapcore.Level
	switch c.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// 配置编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建核心
	var cores []zapcore.Core

	// 添加文件输出
	if c.FilePath != "" {
		fileWriter, err := os.OpenFile(c.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			log.DefaultLogger.Log(log.LevelError, "msg", "failed to open log file", "error", err)
		} else {
			fileCore := zapcore.NewCore(
				zapcore.NewJSONEncoder(encoderConfig),
				zapcore.AddSync(fileWriter),
				level,
			)
			cores = append(cores, fileCore)
		}
	}

	// 添加控制台输出
	if c.Console {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	// 如果没有配置任何输出，使用默认的标准输出
	if len(cores) == 0 {
		return NewDefaultLogger()
	}

	// 创建 zap logger
	core := zapcore.NewTee(cores...)
	zapLogger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	// 使用 kratos zap 适配器
	logger := kratoszap.NewLogger(zapLogger)

	return log.With(logger,
		"service", "goanalysis",
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
	)
}

// NewDefaultLogger 创建默认的日志记录器（用于配置不可用时）
func NewDefaultLogger() log.Logger {
	// 创建基本的控制台配置
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	// 创建 zap logger
	zapLogger, _ := config.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(2),
	)

	// 使用 kratos zap 适配器
	logger := kratoszap.NewLogger(zapLogger)

	return log.With(logger,
		"service", "goanalysis",
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
	)
}
