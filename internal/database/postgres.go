package database

import (
	"fmt"
	"time"

	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresDB struct {
	db             *gorm.DB
	tx             *gorm.DB
	shouldRollback bool
	logger         *zap.SugaredLogger
}

func NewPostgresDB(cfg *config.Config) (*PostgresDB, error) {
	log := utils.NewNamedLogger("database")
	databaseUrl := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable", cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Name, cfg.DB.Password)
	log.Infof("Connecting to the database: %s", databaseUrl)
	db, err := gorm.Open(postgres.Open(databaseUrl), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDb, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDb.SetConnMaxIdleTime(1 * time.Minute)
	sqlDb.SetConnMaxLifetime(10 * time.Minute)
	sqlDb.SetMaxIdleConns(10)
	sqlDb.SetMaxOpenConns(100)
	return &PostgresDB{db: db, logger: log}, nil

}

func (p *PostgresDB) Db() *gorm.DB {
	return p.db
}

func (p *PostgresDB) NewSession() Database {
	return &PostgresDB{db: p.db.Session(&gorm.Session{}), logger: p.logger}
}

func (p *PostgresDB) BeginTransaction() (*gorm.DB, error) {
	if p.tx != nil {
		return p.tx, nil
	}
	tx := p.db.Begin()
	if tx.Error != nil {
		p.logger.Errorf("Failed to start transaction: %s", tx.Error.Error())
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
