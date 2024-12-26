package config

import (
	"os"
	"strconv"

	"github.com/mini-maxit/backend/internal/logger"
	"go.uber.org/zap"
	//"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/internal/logger"
)

const TEST_DB_NAME = "test-maxit"

type Config struct {
	FileStorageUrl string
	DB             DBConfig
	App            AppConfig
	BrokerConfig   BrokerConfig
}

type DBConfig struct {
	Host     string
	Port     uint16
	User     string
	Password string
	Name     string
}

type AppConfig struct {
	Port uint16
}

type BrokerConfig struct {
	// Queue name for sending tasks
	QueueName string
	// Queue name for receiving responses
	ResponseQueueName string
	// RabbitMQ host
	Host string
	// RabbitMQ port
	Port uint16
	// RabbitMQ user
	User string
	// RabbitMQ password
	Password string
}

const (
	DEFAULT_PORT                = "8080"
	DEFAULT_QUEUE_NAME          = "worker_queue"
	DEFAULT_RESPONSE_QUEUE_NAME = "worker_response_queue"
)

func NewConfig() *Config {
	config_logger := logger.NewNamedLogger("config")

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		config_logger.Panic("DB_HOST is not set")
	}
	dbPortStr := os.Getenv("DB_PORT")
	if dbPortStr == "" {
		config_logger.Panic("DB_PORT is not set")
	}
	dbPort := validatePort(dbPortStr, "database", config_logger)
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		config_logger.Panic("DB_USER is not set")
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		config_logger.Warnf("DB_PASSWORD is not set. Using empty password")
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		config_logger.Panic("DB_NAME is not set")
	}

	appPortStr := os.Getenv("APP_PORT")
	if appPortStr == "" {
		config_logger.Warnf("APP_PORT is not set. Using default port %s", DEFAULT_PORT)
		appPortStr = DEFAULT_PORT
	}
	appPort := validatePort(appPortStr, "application", config_logger)

	fileStorageHost := os.Getenv("FILE_STORAGE_HOST")
	if fileStorageHost == "" {
		config_logger.Panic("FILE_STORAGE_HOST is not set")
	}
	fileStoragePortStr := os.Getenv("FILE_STORAGE_PORT")
	if fileStoragePortStr == "" {
		config_logger.Panic("FILE_STORAGE_PORT is not set")
	}
	_ = validatePort(fileStoragePortStr, "file storage", config_logger)

	fileStorageUrl := "http://" + fileStorageHost + ":" + fileStoragePortStr

	queueName := os.Getenv("QUEUE_NAME")
	if queueName == "" {
		config_logger.Warnf("QUEUE_NAME is not set. Using default queue name %s", DEFAULT_QUEUE_NAME)
		queueName = DEFAULT_QUEUE_NAME
	}
	responseQueueName := os.Getenv("RESPONSE_QUEUE_NAME")
	if responseQueueName == "" {
		config_logger.Warnf("RESPONSE_QUEUE_NAME is not set. Using default response queue name %s", DEFAULT_RESPONSE_QUEUE_NAME)
		responseQueueName = DEFAULT_RESPONSE_QUEUE_NAME
	}
	queueHost := os.Getenv("QUEUE_HOST")
	if queueHost == "" {
		config_logger.Panic("QUEUE_HOST is not set")
	}
	queuePortStr := os.Getenv("QUEUE_PORT")
	if queuePortStr == "" {
		config_logger.Panic("QUEUE_PORT is not set")
	}
	queuePort := validatePort(queuePortStr, "broker", config_logger)

	queueUser := os.Getenv("QUEUE_USER")
	if queueUser == "" {
		config_logger.Panic("QUEUE_USER is not set")
	}
	queuePassword := os.Getenv("QUEUE_PASSWORD")
	if queuePassword == "" {
		config_logger.Panic("QUEUE_PASSWORD is not set")
	}

	return &Config{
		DB: DBConfig{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPassword,
			Name:     dbName,
		},
		App: AppConfig{
			Port: appPort,
		},
		BrokerConfig: BrokerConfig{
			QueueName:         queueName,
			ResponseQueueName: responseQueueName,
			Host:              queueHost,
			Port:              queuePort,
			User:              queueUser,
			Password:          queuePassword,
		},
		FileStorageUrl: fileStorageUrl,
	}
}

func validatePort(port string, which string, log *zap.SugaredLogger) uint16 {
	p, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		log.Panicf("invalid %s port number %s", which, port)
	}
	return uint16(p)
}
