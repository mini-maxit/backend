package database

import "gorm.io/gorm"

type Database interface {
	Connect() (*gorm.DB, error) // Returns opened database connection with transaction
	ShouldRollback() bool       // Returns whether the transaction should be rolled back
	Rollback()                  // Sets the transaction to be rolled back after execution finishes
	Commit() error              // Commits the transaction
}
