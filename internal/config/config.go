package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
)

type Config struct {
	FileStorageURL       string
	FileStoragePublicURL string
	DB                   DBConfig
	API                  APIConfig
	Broker               BrokerConfig
	CORS                 CORSConfig
	JWTSecretKey         string
	Dump                 bool
	SignedURLTTLSeconds  uint16
}

type DBConfig struct {
	Host     string
	Port     uint16
	User     string
	Password string
	Name     string
}

type APIConfig struct {
	Port               uint16
	RefreshTokenPath   string
	AccessTokenMinutes uint16
	// CookieSecure sets the Secure flag on the refresh-token cookie. Must be true in production (HTTPS).
	CookieSecure bool
}

type CORSConfig struct {
	// Allowed origins (comma-separated). Use "*" for all origins
	AllowedOrigins string
	// Allow credentials
	AllowCredentials bool
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
	defaultAPIPort                = "8080"
	defaultAPIRefreshTokenPath    = "/api/v1/auth/refresh"
	defaultQueueName              = "worker_queue"
	defaultResponseQueueName      = "worker_response_queue"
	defaultCORSAllowedOrigins     = "http://localhost:3000,http://localhost:5173"
	defaultAccessTokenMinutesStr  = "180"
	defaultSignedURLTTLSecondsStr = "300" // 5 minutes
	trueValue                     = "true"
)

// Environment variable names read by NewConfig.
const (
	envDBHost          = "DB_HOST"
	envDBPort          = "DB_PORT"
	envDBUser          = "DB_USER"
	envDBPassword      = "DB_PASSWORD"
	envDBName          = "DB_NAME"
	envAPPPort         = "APP_PORT"
	envRefreshToken    = "API_REFRESH_TOKEN_PATH"
	envAccessTokenMin  = "JWT_ACCESS_TOKEN_MINUTES"
	envCookieSecure    = "COOKIE_SECURE"
	envFileStorageHost = "FILE_STORAGE_HOST"
	envFileStoragePort = "FILE_STORAGE_PORT"
	envFileStoragePub  = "FILE_STORAGE_PUBLIC_URL"
	envQueueName       = "QUEUE_NAME"
	envResponseQueue   = "RESPONSE_QUEUE_NAME"
	envQueueHost       = "QUEUE_HOST"
	envQueuePort       = "QUEUE_PORT"
	envQueueUser       = "QUEUE_USER"
	envQueuePassword   = "QUEUE_PASSWORD"
	envJWTSecretKey    = "JWT_SECRET_KEY"
	envDump            = "DUMP"
	envCORSOrigins     = "CORS_ALLOWED_ORIGINS"
	envCORSCredentials = "CORS_ALLOW_CREDENTIALS"
	envSignedURLTTL    = "SIGNED_URL_TTL_SECONDS"
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
//   - APP_PORT - application port. Default is 8080
//
//   - FILE_STORAGE_HOST - file storage internal host. Required
//
//   - FILE_STORAGE_PORT - file storage internal port. Required
//
//   - FILE_STORAGE_PUBLIC_URL - public base URL for signed file downloads (e.g. https://host/files). Falls back to internal URL if unset (not reachable by browsers)
//
//   - COOKIE_SECURE - set to "true" to set the Secure flag on the refresh-token cookie. Required in production (HTTPS)
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
//   - JWT_ACCESS_TOKEN_MINUTES - access token lifetime in minutes. Default is 180
//
//   - SIGNED_URL_TTL_SECONDS - time-to-live for signed file URLs in seconds. Default is 300 (5 minutes)
//
//   - LANGUAGES - comma-separated list of languages with their version,
//     e.g. "c:99,c:11,c:18,cpp:11,cpp:14,cpp:17,cpp:20,cpp:23". Default will expand to [DefaultLanguages]
func NewConfig() *Config {
	log := utils.NewNamedLogger("config")

	dbHost := os.Getenv(envDBHost)
	if dbHost == "" {
		log.Warnf(envDBHost+" is not set. Using default value %s", "localhost")
	}
	dbPortStr := os.Getenv(envDBPort)
	if dbPortStr == "" {
		log.Panic(envDBPort + " is not set")
	}
	dbPort := validatePort(dbPortStr, "database", log)
	dbUser := os.Getenv(envDBUser)
	if dbUser == "" {
		log.Panic(envDBUser + " is not set")
	}
	dbPassword := os.Getenv(envDBPassword)
	if dbPassword == "" {
		log.Warnf(envDBPassword + " is not set. Using empty password")
	}
	dbName := os.Getenv(envDBName)
	if dbName == "" {
		log.Panic(envDBName + " is not set")
	}

	appPortStr := os.Getenv(envAPPPort)
	if appPortStr == "" {
		log.Warnf(envAPPPort+" is not set. Using default port %s", defaultAPIPort)
		appPortStr = defaultAPIPort
	}
	appPort := validatePort(appPortStr, "application", log)

	refreshTokenPath := os.Getenv(envRefreshToken)
	if refreshTokenPath == "" {
		log.Warnf(envRefreshToken+" is not set. Using default path %s", defaultAPIRefreshTokenPath)
		refreshTokenPath = defaultAPIRefreshTokenPath
	}

	accessTokenMinutesStr := os.Getenv(envAccessTokenMin)
	if accessTokenMinutesStr == "" {
		log.Warnf(envAccessTokenMin+" is not set. Using default value %s", defaultAccessTokenMinutesStr)
		accessTokenMinutesStr = defaultAccessTokenMinutesStr
	}
	accessTokenMinutesParsed, err := strconv.ParseUint(accessTokenMinutesStr, 10, 16)
	if err != nil {
		log.Panicf("invalid JWT_ACCESS_TOKEN_MINUTES value %s", accessTokenMinutesStr)
	}
	accessTokenMinutes := uint16(accessTokenMinutesParsed)

	cookieSecure := os.Getenv(envCookieSecure) == trueValue

	fileStorageHost := os.Getenv(envFileStorageHost)
	if fileStorageHost == "" {
		log.Panic(envFileStorageHost + " is not set")
	}
	fileStoragePortStr := os.Getenv(envFileStoragePort)
	if fileStoragePortStr == "" {
		log.Panic(envFileStoragePort + " is not set")
	}
	_ = validatePort(fileStoragePortStr, "file storage", log)

	fileStorageURL := "http://" + fileStorageHost + ":" + fileStoragePortStr

	fileStoragePublicURL := strings.TrimSuffix(os.Getenv(envFileStoragePub), "/")
	if fileStoragePublicURL == "" {
		log.Warnf(envFileStoragePub+" is not set. Signed URLs will use internal address %s and will not be reachable by browsers", fileStorageURL)
		fileStoragePublicURL = fileStorageURL
	}

	queueName := os.Getenv(envQueueName)
	if queueName == "" {
		log.Warnf(envQueueName+" is not set. Using default queue name %s", defaultQueueName)
		queueName = defaultQueueName
	}
	responseQueueName := os.Getenv(envResponseQueue)
	if responseQueueName == "" {
		log.Warnf(envResponseQueue+" is not set. Using default response queue name %s", defaultResponseQueueName)
		responseQueueName = defaultResponseQueueName
	}
	queueHost := os.Getenv(envQueueHost)
	if queueHost == "" {
		log.Panic(envQueueHost + " is not set")
	}
	queuePortStr := os.Getenv(envQueuePort)
	if queuePortStr == "" {
		log.Panic(envQueuePort + " is not set")
	}
	queuePort := validatePort(queuePortStr, "broker", log)

	queueUser := os.Getenv(envQueueUser)
	if queueUser == "" {
		log.Panic(envQueueUser + " is not set")
	}
	queuePassword := os.Getenv(envQueuePassword)
	if queuePassword == "" {
		log.Panic(envQueuePassword + " is not set")
	}

	jwtSecretKey := os.Getenv(envJWTSecretKey)
	if jwtSecretKey == "" {
		log.Panic(envJWTSecretKey + " is not set")
	}

	dumpStr := os.Getenv(envDump)
	dump := dumpStr == trueValue

	corsAllowedOrigins := os.Getenv(envCORSOrigins)
	if corsAllowedOrigins == "" {
		log.Warnf(envCORSOrigins+" is not set. Using default value %s", defaultCORSAllowedOrigins)
		corsAllowedOrigins = defaultCORSAllowedOrigins
	}
	corsAllowCredentials := os.Getenv(envCORSCredentials) == trueValue

	if corsAllowCredentials && corsAllowedOrigins == "*" {
		log.Panicf(`CORS_ALLOWED_ORIGINS=* and CORS_ALLOW_CREDENTIALS=true cannot be set at the same time.
More info: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS/Errors/CORSNotSupportingCredentials`)
	}

	signedURLTTLSecondsStr := os.Getenv(envSignedURLTTL)
	if signedURLTTLSecondsStr == "" {
		log.Warnf(envSignedURLTTL+" is not set. Using default value %s", defaultSignedURLTTLSecondsStr)
		signedURLTTLSecondsStr = defaultSignedURLTTLSecondsStr
	}
	signedURLTTLSecondsParsed, err := strconv.ParseUint(signedURLTTLSecondsStr, 10, 16)
	if err != nil {
		log.Panicf("invalid SIGNED_URL_TTL_SECONDS value %s", signedURLTTLSecondsStr)
	}
	signedURLTTLSeconds := uint16(signedURLTTLSecondsParsed)

	return &Config{
		DB: DBConfig{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPassword,
			Name:     dbName,
		},
		API: APIConfig{
			Port:               appPort,
			RefreshTokenPath:   refreshTokenPath,
			AccessTokenMinutes: accessTokenMinutes,
			CookieSecure:       cookieSecure,
		},
		Broker: BrokerConfig{
			QueueName:         queueName,
			ResponseQueueName: responseQueueName,
			Host:              queueHost,
			Port:              queuePort,
			User:              queueUser,
			Password:          queuePassword,
		},
		CORS: CORSConfig{
			AllowedOrigins:   corsAllowedOrigins,
			AllowCredentials: corsAllowCredentials,
		},
		FileStorageURL:       fileStorageURL,
		FileStoragePublicURL: fileStoragePublicURL,
		JWTSecretKey:         jwtSecretKey,
		Dump:                 dump,
		SignedURLTTLSeconds:  signedURLTTLSeconds,
	}
}

func validatePort(port string, which string, log *zap.SugaredLogger) uint16 {
	p, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		log.Panicf("invalid %s port number %s", which, port)
	}
	return uint16(p)
}
