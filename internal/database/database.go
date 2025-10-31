package database

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Database interface {
	BeginTransaction() (*DB, error) // Returns opened database connection with transaction
	NewSession() Database           // Returns a new session
	ShouldRollback() bool           // Returns whether the transaction should be rolled back
	Rollback()                      // Sets the transaction to be rolled back after execution finishes
	Commit() error                  // Commits the transaction
	DB() *DB                        // Returns the database connection wrapped in our DB type
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
