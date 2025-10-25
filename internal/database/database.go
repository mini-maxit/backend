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
}

const SchemaName = "maxit"

var GormConfig = &gorm.Config{
	NamingStrategy: schema.NamingStrategy{
		TablePrefix: fmt.Sprintf("%s.", SchemaName),
	},
}
