package testutils

import (
	"fmt"
	"log"
	"testing"

	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/rabbitmq/amqp091-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewTestConfig() *config.Config {
	return &config.Config{
		DB: config.DBConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			Name:     "test-maxit",
		},
		App: config.AppConfig{
			Port: 8080,
		},
		BrokerConfig: config.BrokerConfig{
			QueueName:         "test_worker_queue",
			ResponseQueueName: "test_worker_response_queue",
			Host:              "localhost",
			Port:              5672,
			User:              "guest",
			Password:          "guest",
		},
	}
}

func NewTestPostgresDB(t *testing.T, cfg *config.Config) database.Database {
	databaseUrl := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable", cfg.DB.Host, cfg.DB.Port, cfg.DB.User, config.TEST_DB_NAME, cfg.DB.Password)
	db, err := gorm.Open(postgres.Open(databaseUrl), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	// db, err := gorm.Open(postgres.Open(databaseUrl), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db connection %v", err)
	}
	// Check if the database has any tables
	var tableNames []string
	err = db.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public'").Scan(&tableNames).Error
	if err != nil {
		t.Fatalf("failed to query tables: %v", err)
	}

	// If the database is not empty, drop all tables
	if len(tableNames) > 0 {
		log.Println("Database is not empty, dropping all tables...")
		err = db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;").Error
		if err != nil {
			t.Fatalf("failed drop tables: %v", err)
		}
		log.Println("All tables dropped successfully.")
	} else {
		log.Println("Database is empty.")
	}
	_, err = repository.NewLanguageRepository(db)
	if err != nil {
		t.Fatalf("failed to create languege repository: %v", err)
	}
	_, err = repository.NewUserRepository(db)
	if err != nil {
		t.Fatalf("failed to create user repository %v", err)
	}
	_, err = repository.NewTaskRepository(db)
	if err != nil {
		t.Fatalf("failed to create task repository %v", err)
	}
	_, err = repository.NewGroupRepository(db)
	if err != nil {
		t.Fatalf("failed to create group repository %v", err)
	}
	_, err = repository.NewSubmissionRepository(db)
	if err != nil {
		t.Fatalf("failed to create submission repository %v", err)
	}
	_, err = repository.NewSubmissionResultRepository(db)
	if err != nil {
		t.Fatalf("failed to create submission result repository %v", err)
	}
	_, err = repository.NewQueueMessageRepository(db)
	if err != nil {
		t.Fatalf("failed to create queue message repository %f", err)
	}
	_, err = repository.NewSessionRepository(db)
	if err != nil {
		t.Fatalf("failed to create session repository %f", err)
	}

	return &database.PostgresDB{Db: db}
}

func NewTestTx(t *testing.T) *gorm.DB {
	cfg := NewTestConfig()
	database := NewTestPostgresDB(t, cfg)
	tx, err := database.Connect()
	if err != nil {
		t.Fatalf("failed to create a new database connection: %v", err)
	}

	return tx
}

func NewTestChannel(t *testing.T) (*amqp091.Connection, *amqp091.Channel) {
	cfg := NewTestConfig()
	conn, err := amqp091.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.BrokerConfig.User, cfg.BrokerConfig.Password, cfg.BrokerConfig.Host, cfg.BrokerConfig.Port))
	if err != nil {
		t.Fatalf("failed to create a new amqp connection: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("failed to create a new amqp channel: %v", err)
	}

	return conn, ch
}
