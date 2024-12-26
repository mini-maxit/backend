package logger

import (
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

var sugar_logger *zap.SugaredLogger
var http_sugar_logger *zap.SugaredLogger

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
	file_core := zapcore.NewCore(
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

	// Combine the cores
	core := zapcore.NewTee(file_core, std_core)

	// Initialize the sugared log for std and file logging
	log := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	sugar_logger = log.Sugar()

	// Initialize the sugared logger for http logging only to file
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
func NewNamedLogger(serviceName string) *zap.SugaredLogger {
	if sugar_logger == nil {
		InitializeLogger()
	}
	return sugar_logger.Named(serviceName)
}
