package config

import (
	"os"
	"strconv"

	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	//"github.com/mini-maxit/backend/internal/config"
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
	log := utils.NewNamedLogger("config")

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		log.Panic("DB_HOST is not set")
	}
	dbPortStr := os.Getenv("DB_PORT")
	if dbPortStr == "" {
		log.Panic("DB_PORT is not set")
	}
	dbPort := validatePort(dbPortStr, "database", log)
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		log.Panic("DB_USER is not set")
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		log.Warnf("DB_PASSWORD is not set. Using empty password")
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		log.Panic("DB_NAME is not set")
	}

	appPortStr := os.Getenv("APP_PORT")
	if appPortStr == "" {
		log.Warnf("APP_PORT is not set. Using default port %s", DEFAULT_PORT)
		appPortStr = DEFAULT_PORT
	}
	appPort := validatePort(appPortStr, "application", log)

	fileStorageHost := os.Getenv("FILE_STORAGE_HOST")
	if fileStorageHost == "" {
		log.Panic("FILE_STORAGE_HOST is not set")
	}
	fileStoragePortStr := os.Getenv("FILE_STORAGE_PORT")
	if fileStoragePortStr == "" {
		log.Panic("FILE_STORAGE_PORT is not set")
	}
	_ = validatePort(fileStoragePortStr, "file storage", log)

	fileStorageUrl := "http://" + fileStorageHost + ":" + fileStoragePortStr

	queueName := os.Getenv("QUEUE_NAME")
	if queueName == "" {
		log.Warnf("QUEUE_NAME is not set. Using default queue name %s", DEFAULT_QUEUE_NAME)
		queueName = DEFAULT_QUEUE_NAME
	}
	responseQueueName := os.Getenv("RESPONSE_QUEUE_NAME")
	if responseQueueName == "" {
		log.Warnf("RESPONSE_QUEUE_NAME is not set. Using default response queue name %s", DEFAULT_RESPONSE_QUEUE_NAME)
		responseQueueName = DEFAULT_RESPONSE_QUEUE_NAME
	}
	queueHost := os.Getenv("QUEUE_HOST")
	if queueHost == "" {
		log.Panic("QUEUE_HOST is not set")
	}
	queuePortStr := os.Getenv("QUEUE_PORT")
	if queuePortStr == "" {
		log.Panic("QUEUE_PORT is not set")
	}
	queuePort := validatePort(queuePortStr, "broker", log)

	queueUser := os.Getenv("QUEUE_USER")
	if queueUser == "" {
		log.Panic("QUEUE_USER is not set")
	}
	queuePassword := os.Getenv("QUEUE_PASSWORD")
	if queuePassword == "" {
		log.Panic("QUEUE_PASSWORD is not set")
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
