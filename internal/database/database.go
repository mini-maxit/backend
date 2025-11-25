package database

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Database interface {
	BeginTransaction() (*gorm.DB, error) // Returns opened database connection with transaction
	NewSession() Database                // Returns a new session
	ShouldRollback() bool                // Returns whether the transaction should be rolled back
	Rollback()                           // Sets the transaction to be rolled back after execution finishes
	Commit() error                       // Commits the transaction
	DB() *gorm.DB                        // Returns the database connection
	GetInstance() *gorm.DB               // Returns the underlying *gorm.DB transaction (for repository use)
	// ResolveTableName(model interface{}) string // Returns the full table name with schema prefix
}

const SchemaName = "maxit"

var GormConfig = &gorm.Config{
	NamingStrategy: schema.NamingStrategy{
		TablePrefix: fmt.Sprintf("%s.", SchemaName),
	},
}

func ResolveTableName(db *gorm.DB, model any) string {
	stmt := &gorm.Statement{DB: db}
	err := stmt.Parse(model)
	if err != nil {
		return ""
	}
	return stmt.Schema.Table
}
