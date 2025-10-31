package testutils

import (
	"context"
	"net/http"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MockDatabase struct {
	invalid bool
	db      *gorm.DB
}

func (m *MockDatabase) BeginTransaction() (database.Database, error) {
	if m.invalid {
		return nil, gorm.ErrInvalidDB
	}
	return &MockDatabase{db: &gorm.DB{}}, nil
}

func (m *MockDatabase) NewSession() database.Database {
	return &MockDatabase{db: &gorm.DB{}}
}

func (m *MockDatabase) Commit() error {
	return nil
}

func (m *MockDatabase) Rollback() {
}

func (m *MockDatabase) ShouldRollback() bool {
	return false
}

func (m *MockDatabase) Invalidate() {
	m.invalid = true
}

func (m *MockDatabase) Validate() {
	m.invalid = false
}

func (m *MockDatabase) GormDB() *gorm.DB {
	if m.db == nil {
		return &gorm.DB{}
	}
	return m.db
}

// Custom methods
func (m *MockDatabase) Join(joinType string, model interface{}, condition string, args ...interface{}) database.Database {
	return m
}

func (m *MockDatabase) ApplyPaginationAndSort(limit, offset int, sortBy string) database.Database {
	return m
}

// GORM forwarding methods
func (m *MockDatabase) Model(value interface{}) database.Database {
	return m
}

func (m *MockDatabase) Where(query interface{}, args ...interface{}) database.Database {
	return m
}

func (m *MockDatabase) Create(value interface{}) database.Database {
	return m
}

func (m *MockDatabase) Save(value interface{}) database.Database {
	return m
}

func (m *MockDatabase) Delete(value interface{}, conds ...interface{}) database.Database {
	return m
}

func (m *MockDatabase) Updates(values interface{}) database.Database {
	return m
}

func (m *MockDatabase) Update(column string, value interface{}) database.Database {
	return m
}

func (m *MockDatabase) First(dest interface{}, conds ...interface{}) database.Database {
	return m
}

func (m *MockDatabase) Take(dest interface{}, conds ...interface{}) database.Database {
	return m
}

func (m *MockDatabase) Find(dest interface{}, conds ...interface{}) database.Database {
	return m
}

func (m *MockDatabase) Count(count *int64) database.Database {
	return m
}

func (m *MockDatabase) Preload(query string, args ...interface{}) database.Database {
	return m
}

func (m *MockDatabase) Joins(query string, args ...interface{}) database.Database {
	return m
}

func (m *MockDatabase) Select(query interface{}, args ...interface{}) database.Database {
	return m
}

func (m *MockDatabase) Group(name string) database.Database {
	return m
}

func (m *MockDatabase) Having(query interface{}, args ...interface{}) database.Database {
	return m
}

func (m *MockDatabase) Order(value interface{}) database.Database {
	return m
}

func (m *MockDatabase) Limit(limit int) database.Database {
	return m
}

func (m *MockDatabase) Offset(offset int) database.Database {
	return m
}

func (m *MockDatabase) Scan(dest interface{}) database.Database {
	return m
}

func (m *MockDatabase) Distinct(args ...interface{}) database.Database {
	return m
}

func (m *MockDatabase) Omit(columns ...string) database.Database {
	return m
}

func (m *MockDatabase) Clauses(conds ...clause.Expression) database.Database {
	return m
}

func (m *MockDatabase) Scopes(funcs ...func(*gorm.DB) *gorm.DB) database.Database {
	return m
}

func (m *MockDatabase) Session(config *gorm.Session) database.Database {
	return m
}

func (m *MockDatabase) Begin() database.Database {
	return m
}

func (m *MockDatabase) Error() error {
	if m.db == nil {
		return nil
	}
	return m.db.Error
}

func (m *MockDatabase) RowsAffected() int64 {
	if m.db == nil {
		return 0
	}
	return m.db.RowsAffected
}

func MockDatabaseMiddleware(next http.Handler, db database.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
