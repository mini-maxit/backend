package utils

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const logPath = "logs/services/log.txt"
const httpLogPath = "logs/http/log.txt"

const (
	timeKey   = "time"
	levelKey  = "level"
	sourceKey = "source"
	msgKey    = "msg"
)

var sugar_logger *zap.SugaredLogger
var http_sugar_logger *zap.SugaredLogger

func InitializeLogger() {
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename: logPath,
		MaxAge:   28,
		Compress: true,
	})

	std_w := zapcore.AddSync(os.Stdout)

	http_w := zapcore.AddSync(&lumberjack.Logger{
		Filename: httpLogPath,
		MaxAge:   1,
		Compress: true,
	})

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        timeKey,
		LevelKey:       levelKey,
		NameKey:        "source",
		MessageKey:     msgKey,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

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

	core := zapcore.NewTee(file_core, std_core)

	log := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	sugar_logger = log.Sugar()

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
