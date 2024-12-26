package database

import (
	"fmt"

	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/internal/logger"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresDB struct {
	Db *gorm.DB
	db_logger *zap.SugaredLogger
}

func NewPostgresDB(cfg *config.Config) (*PostgresDB, error) {
	db_logger := logger.NewNamedLogger("database")
	databaseUrl := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable", cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Name, cfg.DB.Password)
	db_logger.Infof("Connecting to the database: %s", databaseUrl)
	db, err := gorm.Open(postgres.Open(databaseUrl), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &PostgresDB{Db: db, db_logger: db_logger}, nil

}

func (p *PostgresDB) Connect() *gorm.DB {
	return p.Db
}
