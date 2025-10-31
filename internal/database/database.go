package database

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type Database interface {
	// Transaction management methods
	BeginTransaction() (Database, error) // Returns opened database connection with transaction
	NewSession() Database                // Returns a new session
	ShouldRollback() bool                // Returns whether the transaction should be rolled back
	Rollback()                           // Sets the transaction to be rolled back after execution finishes
	Commit() error                       // Commits the transaction

	// Custom methods
	Join(joinType string, model interface{}, condition string, args ...interface{}) Database
	ApplyPaginationAndSort(limit, offset int, sortBy string) Database

	// GORM forwarding methods
	Model(value interface{}) Database
	Where(query interface{}, args ...interface{}) Database
	Create(value interface{}) Database
	Save(value interface{}) Database
	Delete(value interface{}, conds ...interface{}) Database
	Updates(values interface{}) Database
	Update(column string, value interface{}) Database
	First(dest interface{}, conds ...interface{}) Database
	Take(dest interface{}, conds ...interface{}) Database
	Find(dest interface{}, conds ...interface{}) Database
	Count(count *int64) Database
	Preload(query string, args ...interface{}) Database
	Joins(query string, args ...interface{}) Database
	Select(query interface{}, args ...interface{}) Database
	Group(name string) Database
	Having(query interface{}, args ...interface{}) Database
	Order(value interface{}) Database
	Limit(limit int) Database
	Offset(offset int) Database
	Scan(dest interface{}) Database
	Distinct(args ...interface{}) Database
	Omit(columns ...string) Database
	Clauses(conds ...clause.Expression) Database
	Scopes(funcs ...func(*gorm.DB) *gorm.DB) Database
	Session(config *gorm.Session) Database
	Begin() Database

	// Error handling
	Error() error
	RowsAffected() int64

	// Internal method to get underlying gorm.DB
	GormDB() *gorm.DB
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
