package config

import (
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
)

type Config struct {
	FileStorageUrl string
	DB             DBConfig
	Api            ApiConfig
	BrokerConfig   BrokerConfig
	Dump           bool

	EnabledLanguages []schemas.LanguageConfig
}

type DBConfig struct {
	Host     string
	Port     uint16
	User     string
	Password string
	Name     string
}

type ApiConfig struct {
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
	defaultApiPort           = "8080"
	defaultQueueName         = "worker_queue"
	defaultResponseQueueName = "worker_response_queue"
)

// DefaultLanguages is a list of languages that is enabled by default
var DefaultLanguages = []schemas.LanguageConfig{
	{
		Type:    models.LangTypeC,
		Version: "99",
	},
	{
		Type:    models.LangTypeC,
		Version: "11",
	},
	{
		Type:    models.LangTypeC,
		Version: "18",
	},
	{
		Type:    models.LangTypeCPP,
		Version: "11",
	},
	{
		Type:    models.LangTypeCPP,
		Version: "14",
	},
	{
		Type:    models.LangTypeCPP,
		Version: "17",
	},
	{
		Type:    models.LangTypeCPP,
		Version: "20",
	},
	{
		Type:    models.LangTypeCPP,
		Version: "23",
	},
}

// AvailableLanguages is a list of languages that is acrively supported by the system and can be used if enabled.
var AvailableLanguages = []schemas.LanguageConfig{
	{
		Type:    models.LangTypeC,
		Version: "99",
	},
	{
		Type:    models.LangTypeC,
		Version: "11",
	},
	{
		Type:    models.LangTypeC,
		Version: "18",
	},
	{
		Type:    models.LangTypeCPP,
		Version: "11",
	},
	{
		Type:    models.LangTypeCPP,
		Version: "14",
	},
	{
		Type:    models.LangTypeCPP,
		Version: "17",
	},
	{
		Type:    models.LangTypeCPP,
		Version: "20",
	},
	{
		Type:    models.LangTypeCPP,
		Version: "23",
	},
}

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
//   - LANGUAGES - comma-separated list of languages with their version, e.g. "c:99,c:11,c:18,cpp:11,cpp:14,cpp:17,cpp:20,cpp:23". Default will exapnd to [DefaultLanguages]
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
		log.Warnf("API_PORT is not set. Using default port %s", defaultApiPort)
		appPortStr = defaultApiPort
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
		Api: ApiConfig{
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
		FileStorageUrl:   fileStorageUrl,
		Dump:             dump,
		EnabledLanguages: parseLanguages(os.Getenv("LANGUAGES"), log),
	}
}

func validatePort(port string, which string, log *zap.SugaredLogger) uint16 {
	p, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		log.Panicf("invalid %s port number %s", which, port)
	}
	return uint16(p)
}

func parseLanguages(input string, log *zap.SugaredLogger) []schemas.LanguageConfig {
	if input == "" {
		log.Warn("LANGUAGES is not set. Using default languages")
		return DefaultLanguages
	}
	langs := make([]schemas.LanguageConfig, 0)
	languages := strings.Split(input, ",")
	for _, lang := range languages {
		parts := strings.Split(lang, ":")
		if len(parts) != 2 {
			log.Panicf("invalid language format in config: %s. For available options refer to documentation", lang)
		}
		language := schemas.LanguageConfig{Type: models.LanguageType(parts[0]), Version: parts[1]}
		if !slices.Contains(AvailableLanguages, language) {
			log.Panicf("language %s is not available. Available languages: %v, for more refer to documentation", lang, AvailableLanguages)
		}
		langs = append(langs, language)
	}

	return langs
}
