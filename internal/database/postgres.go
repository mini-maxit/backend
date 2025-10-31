package database

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostgresDB struct {
	db             *gorm.DB
	tx             *gorm.DB
	shouldRollback bool
	logger         *zap.SugaredLogger
}

func NewPostgresDB(cfg *config.Config) (*PostgresDB, error) {
	log := utils.NewNamedLogger("database")
	databaseURL := fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.User,
		cfg.DB.Name,
		cfg.DB.Password,
	)
	log.Infof("Connecting to the database: %s", databaseURL)
	db, err := gorm.Open(postgres.Open(databaseURL), GormConfig)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetConnMaxIdleTime(1 * time.Minute)
	sqlDB.SetConnMaxLifetime(10 * time.Minute)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	return &PostgresDB{db: db, logger: log}, nil
}

// GormDB returns the underlying gorm.DB instance
func (p *PostgresDB) GormDB() *gorm.DB {
	return p.db
}

func (p *PostgresDB) NewSession() Database {
	return &PostgresDB{db: p.db.Session(&gorm.Session{}), logger: p.logger}
}

func (p *PostgresDB) BeginTransaction() (Database, error) {
	if p.tx != nil {
		return &PostgresDB{db: p.tx, tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}, nil
	}
	tx := p.db.Begin()
	if tx.Error != nil {
		p.logger.Errorf("Failed to start transaction: %s", tx.Error.Error())
		return nil, tx.Error
	}
	p.tx = tx
	return &PostgresDB{db: tx, tx: tx, shouldRollback: p.shouldRollback, logger: p.logger}, nil
}

func (p *PostgresDB) ShouldRollback() bool {
	return p.shouldRollback
}

func (p *PostgresDB) Rollback() {
	p.shouldRollback = true
}

func (p *PostgresDB) Commit() error {
	if p.tx == nil {
		return errors.New("no transaction to commit to")
	}
	p.shouldRollback = false
	p.tx.Commit()
	if p.tx.Error != nil {
		return p.tx.Error
	}
	p.tx = nil
	return nil
}

// Custom methods

// Join performs a JOIN operation with automatic schema prefixing for the joined table
func (p *PostgresDB) Join(joinType string, model interface{}, condition string, args ...interface{}) Database {
	tableName := ResolveTableName(p.db, model)
	joinClause := fmt.Sprintf("%s %s ON %s", joinType, tableName, condition)
	return &PostgresDB{db: p.db.Joins(joinClause, args...), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

// ApplyPaginationAndSort applies pagination and sort to the query
func (p *PostgresDB) ApplyPaginationAndSort(limit, offset int, sortBy string) Database {
	db := p.Limit(limit).Offset(offset)

	if sortBy != "" {
		sortFields := strings.Split(sortBy, ",")
		for _, sortField := range sortFields {
			sortFieldParts := strings.Split(sortField, ":")
			db = db.Order(sortFieldParts[0] + " " + sortFieldParts[1])
		}
	}

	return db
}

// GORM forwarding methods

func (p *PostgresDB) Model(value interface{}) Database {
	return &PostgresDB{db: p.db.Model(value), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Where(query interface{}, args ...interface{}) Database {
	return &PostgresDB{db: p.db.Where(query, args...), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Create(value interface{}) Database {
	return &PostgresDB{db: p.db.Create(value), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Save(value interface{}) Database {
	return &PostgresDB{db: p.db.Save(value), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Delete(value interface{}, conds ...interface{}) Database {
	return &PostgresDB{db: p.db.Delete(value, conds...), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Updates(values interface{}) Database {
	return &PostgresDB{db: p.db.Updates(values), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Update(column string, value interface{}) Database {
	return &PostgresDB{db: p.db.Update(column, value), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) First(dest interface{}, conds ...interface{}) Database {
	return &PostgresDB{db: p.db.First(dest, conds...), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Take(dest interface{}, conds ...interface{}) Database {
	return &PostgresDB{db: p.db.Take(dest, conds...), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Find(dest interface{}, conds ...interface{}) Database {
	return &PostgresDB{db: p.db.Find(dest, conds...), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Count(count *int64) Database {
	return &PostgresDB{db: p.db.Count(count), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Preload(query string, args ...interface{}) Database {
	return &PostgresDB{db: p.db.Preload(query, args...), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Joins(query string, args ...interface{}) Database {
	return &PostgresDB{db: p.db.Joins(query, args...), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Select(query interface{}, args ...interface{}) Database {
	return &PostgresDB{db: p.db.Select(query, args...), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Group(name string) Database {
	return &PostgresDB{db: p.db.Group(name), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Having(query interface{}, args ...interface{}) Database {
	return &PostgresDB{db: p.db.Having(query, args...), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Order(value interface{}) Database {
	return &PostgresDB{db: p.db.Order(value), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Limit(limit int) Database {
	return &PostgresDB{db: p.db.Limit(limit), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Offset(offset int) Database {
	return &PostgresDB{db: p.db.Offset(offset), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Scan(dest interface{}) Database {
	return &PostgresDB{db: p.db.Scan(dest), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Distinct(args ...interface{}) Database {
	return &PostgresDB{db: p.db.Distinct(args...), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Omit(columns ...string) Database {
	return &PostgresDB{db: p.db.Omit(columns...), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Clauses(conds ...clause.Expression) Database {
	return &PostgresDB{db: p.db.Clauses(conds...), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Scopes(funcs ...func(*gorm.DB) *gorm.DB) Database {
	return &PostgresDB{db: p.db.Scopes(funcs...), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Session(config *gorm.Session) Database {
	return &PostgresDB{db: p.db.Session(config), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

func (p *PostgresDB) Begin() Database {
	return &PostgresDB{db: p.db.Begin(), tx: p.tx, shouldRollback: p.shouldRollback, logger: p.logger}
}

// Error handling methods

func (p *PostgresDB) Error() error {
	return p.db.Error
}

func (p *PostgresDB) RowsAffected() int64 {
	return p.db.RowsAffected
}
