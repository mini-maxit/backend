package database

import (
	"fmt"

	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/internal/logger"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresDB struct {
	Db             *gorm.DB
	tx             *gorm.DB
	shouldRollback bool
	db_logger      *logger.ServiceLogger
}

func NewPostgresDB(cfg *config.Config) (*PostgresDB, error) {
	db_logger := logger.NewNamedLogger("database")
	databaseUrl := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable", cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Name, cfg.DB.Password)
	logger.Log(&db_logger, fmt.Sprintf("Connecting to the database: %s", databaseUrl), "", logger.Info)
	db, err := gorm.Open(postgres.Open(databaseUrl), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &PostgresDB{Db: db, db_logger: &db_logger}, nil
}

func (p *PostgresDB) Connect() (*gorm.DB, error) {
	if p.tx != nil {
		return p.tx, nil
	}
	tx := p.Db.Begin()
	if tx.Error != nil {
		logrus.Errorf("Failed to start transaction: %s", tx.Error.Error())
		return nil, tx.Error
	}
	p.tx = tx
	return tx, nil
}

func (p *PostgresDB) ShouldRollback() bool {
	return p.shouldRollback
}

func (p *PostgresDB) Rollback() {
	p.shouldRollback = true
}

func (p *PostgresDB) Commit() error {
	if p.tx == nil {
		return fmt.Errorf("no transaction to commit to")
	}
	p.shouldRollback = false
	p.tx.Commit()
	if p.tx.Error != nil {
		return p.tx.Error
	}
	p.tx = nil
	return nil
}

func (p *PostgresDB) InvalidateTx() {
	p.shouldRollback = false
	if p.tx != nil {
		p.tx.Rollback()
	}
	p.tx = nil
}
