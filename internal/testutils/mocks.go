package testutils

import (
	"context"
	"net/http"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/database"
	"gorm.io/gorm"
)

type MockDatabase struct {
	invalid bool
}

func (db *MockDatabase) BeginTransaction() (*database.DB, error) {
	if db.invalid {
		return nil, gorm.ErrInvalidDB
	}
	return database.NewDB(&gorm.DB{}), nil
}

func (db *MockDatabase) NewSession() database.Database {
	return &MockDatabase{}
}

func (db *MockDatabase) Commit() error {
	return nil
}

func (db *MockDatabase) Rollback() {
}

func (db *MockDatabase) ShouldRollback() bool {
	return false
}

func (db *MockDatabase) Invalidate() {
	db.invalid = true
}

func (db *MockDatabase) Validate() {
	db.invalid = false
}

func (db *MockDatabase) DB() *database.DB {
	return database.NewDB(&gorm.DB{})
}

func MockDatabaseMiddleware(next http.Handler, db database.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
