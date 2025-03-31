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

var sugarLogger *zap.SugaredLogger
var httpSugarLogger *zap.SugaredLogger

// InitializeLogger initializes the logger.
func InitializeLogger() {
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename: logPath,
		MaxAge:   28,
		Compress: true,
	})

	stdWriter := zapcore.AddSync(os.Stdout)

	httpWriter := zapcore.AddSync(&lumberjack.Logger{
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

	fileCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		w,
		zap.InfoLevel,
	)

	stdCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		stdWriter,
		zap.InfoLevel,
	)

	httpCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		httpWriter,
		zap.InfoLevel,
	)

	core := zapcore.NewTee(fileCore, stdCore)

	log := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	sugarLogger = log.Sugar()

	httpLogger := zap.New(httpCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	httpSugarLogger = httpLogger.Sugar()
}

// NewHTTPLogger a new SugaredLogger.
func NewHTTPLogger() *zap.SugaredLogger {
	if httpSugarLogger == nil {
		InitializeLogger()
	}
	return httpSugarLogger.Named("http")
}

// NewNamedLogger creates a new named SugaredLogger for a given service.
func NewNamedLogger(serviceName string) *zap.SugaredLogger {
	if sugarLogger == nil {
		InitializeLogger()
	}
	return sugarLogger.Named(serviceName)
}
