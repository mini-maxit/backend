package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const logPath = "./logger/logs/services/log.txt"
const httpLogPath = "./logger/logs/http/log.txt"

const (
	timeKey   = "time"
	levelKey  = "level"
	sourceKey = "source"
	msgKey    = "msg"
)

type LogType int

const (
	Info LogType = iota
	Error
	Warn
	Panic
)

var sugar_logger *zap.SugaredLogger
var std_sugar_logger *zap.SugaredLogger
var http_sugar_logger *zap.SugaredLogger

type ServiceLogger struct {
	file_logger *zap.SugaredLogger
	std_logger *zap.SugaredLogger
}

// InitializeLogger sets up Zap with a custom configuration and initializes the SugaredLogger
func InitializeLogger() {
	// Configure log rotation with lumberjack
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename: logPath,
		MaxAge:   1,
		Compress: true,
	})

	std_w := zapcore.AddSync(os.Stdout)

	http_w := zapcore.AddSync(&lumberjack.Logger{
		Filename: httpLogPath,
		MaxAge:   1,
		Compress: true,
	})

	// Encoder configuration for Console format
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        timeKey,
		LevelKey:       levelKey,
		NameKey:        "source",
		MessageKey:     msgKey,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	// Create the core
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		w,
		zap.InfoLevel,
	)

	std_core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		std_w,
		zap.InfoLevel,
	)

	http_core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		http_w,
		zap.InfoLevel,
	)

	// Initialize the sugared logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	sugar_logger = logger.Sugar()

	stdLogger := zap.New(std_core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	std_sugar_logger = stdLogger.Sugar()

	httpLogger := zap.New(http_core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	http_sugar_logger = httpLogger.Sugar()
}

func NewHttpLogger() *zap.SugaredLogger {
	if http_sugar_logger == nil {
		InitializeLogger()
	}
	return http_sugar_logger.Named("http")
}

// NewNamedLogger creates a new named SugaredLogger for a given service
func NewNamedLogger(serviceName string) ServiceLogger {
	if sugar_logger == nil || std_sugar_logger == nil {
		InitializeLogger()
	}
	return ServiceLogger{sugar_logger.Named(serviceName), std_sugar_logger.Named(serviceName)}
}

func Log(logger *ServiceLogger, log_message, error_message string, log_type LogType) {
	message := fmt.Sprintf("%s %s", log_message, error_message)
	switch log_type {
	case Info:
		logger.file_logger.Info(message)
		logger.std_logger.Info(message)
	case Error:
		logger.file_logger.Error(message)
		logger.std_logger.Error(message)
	case Warn:
		logger.file_logger.Warn(message)
		logger.std_logger.Warn(message)
	case Panic:
		logger.file_logger.Panic(message)
		logger.std_logger.Panic(message)
	}
}
