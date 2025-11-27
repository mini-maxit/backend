package testutils

import (
	"github.com/mini-maxit/backend/internal/database"
	"gorm.io/gorm"
)

type MockDatabase struct {
	invalid bool
}

func (db MockDatabase) BeginTransaction() (*gorm.DB, error) {
	if db.invalid {
		return nil, gorm.ErrInvalidDB
	}
	return &gorm.DB{}, nil
}

func (db MockDatabase) NewSession() database.Database {
	return &MockDatabase{}
}

func (db MockDatabase) Commit() error {
	return nil
}

func (db MockDatabase) Rollback() {
}

func (db MockDatabase) ShouldRollback() bool {
	return false
}

func (db *MockDatabase) Invalidate() {
	db.invalid = true
}

func (db *MockDatabase) Validate() {
	db.invalid = false
}

func (db MockDatabase) DB() *gorm.DB {
	return &gorm.DB{}
}

func (db MockDatabase) GetInstance() *gorm.DB {
	return &gorm.DB{}
}

func (db MockDatabase) ResolveTableName(model interface{}) string {
	return "mock_table"
}
