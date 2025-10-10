package config

import (
	"os"
	"strconv"

	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
)

type Config struct {
	FileStorageURL string
	DB             DBConfig
	API            APIConfig
	Broker         BrokerConfig
	JWTSecretKey   string
	Dump           bool
}

type DBConfig struct {
	Host     string
	Port     uint16
	User     string
	Password string
	Name     string
}

type APIConfig struct {
	Port             uint16
	RefreshTokenPath string
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
	defaultAPIPort             = "8080"
	defaultAPIRefreshTokenPath = "/api/v1/auth/refresh"
	defaultQueueName           = "worker_queue"
	defaultResponseQueueName   = "worker_response_queue"
)

// NewConfig creates new Config instance
//
// It reads environment variables and returns Config instance. Available environment variables:
//
//   - DB_HOST database host. Required
//
//   - DB_PORT - database port. Required
//
//   - DB_USER - database user. Required
//
//   - DB_PASSWORD - database password. Required
//
//   - DB_NAME - database name. Required
//
//   - API_PORT - application port. Default is 8080
//
//   - FILE_STORAGE_HOST - file storage host. Required
//
//   - QUEUE_NAME - queue name for sending tasks. Default is "worker_queue"
//
//   - RESPONSE_QUEUE_NAME - queue name for receiving responses. Default is "worker_response_queue"
//
//   - QUEUE_HOST - broker host. Required
//
//   - QUEUE_PORT - broker port. Required
//
//   - QUEUE_USER - broker user. Required
//
//   - QUEUE_PASSWORD - broker password. Required
//
//   - JWT_SECRET_KEY - secret key for JWT token signing. Required
//
//   - LANGUAGES - comma-separated list of languages with their version,
//     e.g. "c:99,c:11,c:18,cpp:11,cpp:14,cpp:17,cpp:20,cpp:23". Default will expand to [DefaultLanguages]
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
		log.Warnf("API_PORT is not set. Using default port %s", defaultAPIPort)
		appPortStr = defaultAPIPort
	}
	appPort := validatePort(appPortStr, "application", log)

	refreshTokenPath := os.Getenv("API_REFRESH_TOKEN_PATH")
	if refreshTokenPath == "" {
		log.Warnf("API_REFRESH_TOKEN_PATH is not set. Using default path %s", defaultAPIRefreshTokenPath)
		refreshTokenPath = defaultAPIRefreshTokenPath
	}

	fileStorageHost := os.Getenv("FILE_STORAGE_HOST")
	if fileStorageHost == "" {
		log.Panic("FILE_STORAGE_HOST is not set")
	}
	fileStoragePortStr := os.Getenv("FILE_STORAGE_PORT")
	if fileStoragePortStr == "" {
		log.Panic("FILE_STORAGE_PORT is not set")
	}
	_ = validatePort(fileStoragePortStr, "file storage", log)

	fileStorageURL := "http://" + fileStorageHost + ":" + fileStoragePortStr

	queueName := os.Getenv("QUEUE_NAME")
	if queueName == "" {
		log.Warnf("QUEUE_NAME is not set. Using default queue name %s", defaultQueueName)
		queueName = defaultQueueName
	}
	responseQueueName := os.Getenv("RESPONSE_QUEUE_NAME")
	if responseQueueName == "" {
		log.Warnf("RESPONSE_QUEUE_NAME is not set. Using default response queue name %s", defaultResponseQueueName)
		responseQueueName = defaultResponseQueueName
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

	jwtSecretKey := os.Getenv("JWT_SECRET_KEY")
	if jwtSecretKey == "" {
		log.Panic("JWT_SECRET_KEY is not set")
	}

	dumpStr := os.Getenv("DUMP")
	dump := dumpStr == "true"

	return &Config{
		DB: DBConfig{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPassword,
			Name:     dbName,
		},
		API: APIConfig{
			Port:             appPort,
			RefreshTokenPath: refreshTokenPath,
		},
		Broker: BrokerConfig{
			QueueName:         queueName,
			ResponseQueueName: responseQueueName,
			Host:              queueHost,
			Port:              queuePort,
			User:              queueUser,
			Password:          queuePassword,
		},
		FileStorageURL: fileStorageURL,
		JWTSecretKey:   jwtSecretKey,
		Dump:           dump,
	}
}

func validatePort(port string, which string, log *zap.SugaredLogger) uint16 {
	p, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		log.Panicf("invalid %s port number %s", which, port)
	}
	return uint16(p)
}
