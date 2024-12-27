package testutils

import (
	"fmt"
	"log"
	"testing"

	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/internal/database"
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
			QueueName:         "worker_queue",
			ResponseQueueName: "worker_response_queue",
			Host:              "localhost",
			Port:              5672,
			User:              "guest",
			Password:          "guest",
		},
	}
}

func NewTestPostgresDB(cfg *config.Config) (database.Database, error) {
	databaseUrl := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable", cfg.DB.Host, cfg.DB.Port, cfg.DB.User, config.TEST_DB_NAME, cfg.DB.Password)
	db, err := gorm.Open(postgres.Open(databaseUrl), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	// db, err := gorm.Open(postgres.Open(databaseUrl), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	// Check if the database has any tables
	var tableNames []string
	err = db.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public'").Scan(&tableNames).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}

	// If the database is not empty, drop all tables
	if len(tableNames) > 0 {
		log.Println("Database is not empty, dropping all tables...")
		err = db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;").Error
		if err != nil {
			return nil, fmt.Errorf("failed to drop all tables: %w", err)
		}
		log.Println("All tables dropped successfully.")
	} else {
		log.Println("Database is empty.")
	}
	return &database.PostgresDB{Db: db}, nil
}

func NewTestTx(t *testing.T) *gorm.DB {
	cfg := NewTestConfig()
	database, err := NewTestPostgresDB(cfg)
	if err != nil {
		t.Fatalf("failed to create a new database: %v", err)
	}

	tx, err := database.Connect()
	if err != nil {
		t.Fatalf("failed to create a new database connection: %v", err)
	}

	return tx
}
