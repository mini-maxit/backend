package database

import "gorm.io/gorm"

type Database interface {
	DB() *gorm.DB               // Returns the database connection
	Connect() (*gorm.DB, error) // Returns opened database connection with transaction
	NewSession() Database       // Returns a new session
	ShouldRollback() bool       // Returns whether the transaction should be rolled back
	Rollback()                  // Sets the transaction to be rolled back after execution finishes
	Commit() error              // Commits the transaction
	InvalidateTx()              // Invalidates the transaction
}
