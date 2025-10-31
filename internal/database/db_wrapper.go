package database

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DB is a wrapper around gorm.DB that provides additional functionality
// such as automatic schema prefixing in JOIN operations
type DB struct {
	db *gorm.DB
}

// NewDB creates a new DB wrapper around a gorm.DB instance
func NewDB(gormDB *gorm.DB) *DB {
	return &DB{db: gormDB}
}

// GormDB returns the underlying gorm.DB instance
func (d *DB) GormDB() *gorm.DB {
	return d.db
}

// Join performs a JOIN operation with automatic schema prefixing for the joined table
func (d *DB) Join(joinType string, model interface{}, condition string, args ...interface{}) *DB {
	tableName := ResolveTableName(d.db, model)
	joinClause := fmt.Sprintf("%s %s ON %s", joinType, tableName, condition)
	return &DB{db: d.db.Joins(joinClause, args...)}
}

// ApplyPaginationAndSort applies pagination and sort to the query.
//
// Values received are guaranteed to be valid by middleware, so no error checking is needed.
func (d *DB) ApplyPaginationAndSort(limit, offset int, sortBy string) *DB {
	db := d.Limit(limit).Offset(offset)

	if sortBy != "" {
		sortFields := strings.Split(sortBy, ",")
		for _, sortField := range sortFields {
			sortFieldParts := strings.Split(sortField, ":")
			db = db.Order(sortFieldParts[0] + " " + sortFieldParts[1])
		}
	}

	return db
}

// Model specifies the model for subsequent operations
func (d *DB) Model(value interface{}) *DB {
	return &DB{db: d.db.Model(value)}
}

// Where adds a WHERE clause to the query
func (d *DB) Where(query interface{}, args ...interface{}) *DB {
	return &DB{db: d.db.Where(query, args...)}
}

// Create inserts a new record into the database
func (d *DB) Create(value interface{}) *DB {
	return &DB{db: d.db.Create(value)}
}

// Save updates all fields of a record
func (d *DB) Save(value interface{}) *DB {
	return &DB{db: d.db.Save(value)}
}

// Delete soft deletes a record
func (d *DB) Delete(value interface{}, conds ...interface{}) *DB {
	return &DB{db: d.db.Delete(value, conds...)}
}

// Updates updates the specified fields
func (d *DB) Updates(values interface{}) *DB {
	return &DB{db: d.db.Updates(values)}
}

// Update updates a single field
func (d *DB) Update(column string, value interface{}) *DB {
	return &DB{db: d.db.Update(column, value)}
}

// First finds the first record ordered by primary key
func (d *DB) First(dest interface{}, conds ...interface{}) *DB {
	return &DB{db: d.db.First(dest, conds...)}
}

// Take finds the first record without ordering
func (d *DB) Take(dest interface{}, conds ...interface{}) *DB {
	return &DB{db: d.db.Take(dest, conds...)}
}

// Find finds all records matching the conditions
func (d *DB) Find(dest interface{}, conds ...interface{}) *DB {
	return &DB{db: d.db.Find(dest, conds...)}
}

// Count counts the number of records
func (d *DB) Count(count *int64) *DB {
	return &DB{db: d.db.Count(count)}
}

// Preload preloads associations
func (d *DB) Preload(query string, args ...interface{}) *DB {
	return &DB{db: d.db.Preload(query, args...)}
}

// Joins specifies joins with a raw string
func (d *DB) Joins(query string, args ...interface{}) *DB {
	return &DB{db: d.db.Joins(query, args...)}
}

// Select specifies fields to retrieve
func (d *DB) Select(query interface{}, args ...interface{}) *DB {
	return &DB{db: d.db.Select(query, args...)}
}

// Group specifies the GROUP BY clause
func (d *DB) Group(name string) *DB {
	return &DB{db: d.db.Group(name)}
}

// Having specifies the HAVING clause
func (d *DB) Having(query interface{}, args ...interface{}) *DB {
	return &DB{db: d.db.Having(query, args...)}
}

// Order specifies the ORDER BY clause
func (d *DB) Order(value interface{}) *DB {
	return &DB{db: d.db.Order(value)}
}

// Limit specifies the maximum number of records to retrieve
func (d *DB) Limit(limit int) *DB {
	return &DB{db: d.db.Limit(limit)}
}

// Offset specifies the number of records to skip
func (d *DB) Offset(offset int) *DB {
	return &DB{db: d.db.Offset(offset)}
}

// Scan scans results into a destination
func (d *DB) Scan(dest interface{}) *DB {
	return &DB{db: d.db.Scan(dest)}
}

// Distinct specifies distinct fields
func (d *DB) Distinct(args ...interface{}) *DB {
	return &DB{db: d.db.Distinct(args...)}
}

// Omit specifies fields to omit when creating/updating
func (d *DB) Omit(columns ...string) *DB {
	return &DB{db: d.db.Omit(columns...)}
}

// Clauses adds clauses to the query
func (d *DB) Clauses(conds ...clause.Expression) *DB {
	return &DB{db: d.db.Clauses(conds...)}
}

// Scopes applies query scopes
func (d *DB) Scopes(funcs ...func(*gorm.DB) *gorm.DB) *DB {
	return &DB{db: d.db.Scopes(funcs...)}
}

// Session creates a new session with the given config
func (d *DB) Session(config *gorm.Session) *DB {
	return &DB{db: d.db.Session(config)}
}

// Begin starts a transaction
func (d *DB) Begin() *DB {
	return &DB{db: d.db.Begin()}
}

// Commit commits the transaction
func (d *DB) Commit() *DB {
	return &DB{db: d.db.Commit()}
}

// Rollback rolls back the transaction
func (d *DB) Rollback() *DB {
	return &DB{db: d.db.Rollback()}
}

// Error returns any error that occurred during query execution
func (d *DB) Error() error {
	return d.db.Error
}

// RowsAffected returns the number of rows affected by the query
func (d *DB) RowsAffected() int64 {
	return d.db.RowsAffected
}
