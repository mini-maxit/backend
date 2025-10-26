package database

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SchemaPlugin automatically adds schema prefixes to table names
type SchemaPlugin struct {
	tableNames []string
}

// Name returns the plugin name
func (sp *SchemaPlugin) Name() string {
	return "schema_plugin"
}

// Initialize registers the plugin callbacks and extracts table names from models
func (sp *SchemaPlugin) Initialize(db *gorm.DB) error {
	// Extract table names dynamically from GORM models
	sp.extractTableNames(db)

	// Register callback to process joins before query execution
	return db.Callback().Query().Before("gorm:query").Register("schema_plugin:auto_prefix", sp.autoPrefix)
}

// extractTableNames dynamically gets table names from registered GORM models
func (sp *SchemaPlugin) extractTableNames(db *gorm.DB) {
	sp.tableNames = make([]string, 0, len(models.AllModels))

	for _, model := range models.AllModels {
		stmt := &gorm.Statement{DB: db}
		err := stmt.Parse(model)
		if err == nil {
			// Get the table name without schema prefix
			tableName := stmt.Schema.Table
			// Remove schema prefix if it exists
			if strings.Contains(tableName, ".") {
				parts := strings.Split(tableName, ".")
				tableName = parts[len(parts)-1]
			}
			sp.tableNames = append(sp.tableNames, tableName)
		}
	}
}

// autoPrefix automatically adds schema prefix to table names surrounded by spaces
func (sp *SchemaPlugin) autoPrefix(db *gorm.DB) {
	if db.Statement.SQL.Len() == 0 {
		return
	}

	sql := db.Statement.SQL.String()

	for _, tableName := range sp.tableNames {
		schemaTable := SchemaName + "." + tableName

		// Skip if already has schema prefix
		if strings.Contains(sql, schemaTable) {
			continue
		}

		// Look for table name surrounded by spaces: " table_name "
		pattern := " " + tableName + " "
		replacement := " " + schemaTable + " "

		sql = strings.ReplaceAll(sql, pattern, replacement)
	}

	// Update the SQL in the statement
	db.Statement.SQL.Reset()
	db.Statement.SQL.WriteString(sql)
}

// NewSchemaPlugin creates a new instance of the schema plugin
func NewSchemaPlugin() *SchemaPlugin {
	return &SchemaPlugin{}
}

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

	// Register the schema plugin to automatically handle schema prefixes
	if err := db.Use(NewSchemaPlugin()); err != nil {
		log.Errorf("Failed to register schema plugin: %s", err.Error())
		return nil, err
	}
	log.Info("Schema plugin registered successfully")

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

func (p *PostgresDB) DB() *gorm.DB {
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
