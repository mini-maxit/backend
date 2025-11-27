package httputils

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mini-maxit/backend/internal/database"
)

type QueryError struct {
	Filed  string
	Detail string
}

func (e QueryError) Error() string {
	return fmt.Sprintf("Query error: %s: %s", e.Filed, e.Detail)
}

const MultipleQueryValues = "Multiple values for query parameter"

const DefaultPaginationLimitStr = "10"
const DefaultPaginationOffsetStr = "0"
const DefaultSortOrderField = "id:desc"

type ContextKey string

const (
	// TokenKey is the key used to store the JWT token in the context.
	TokenKey ContextKey = "token"
	// UserIDKey is the key used to store the user ID in the context.
	UserIDKey ContextKey = "userID"
	// UserKey is the key used to store the user in the context.
	UserKey ContextKey = "user"
	// DatabaseKey is the key used to store the database connection in the context.
	DatabaseKey ContextKey = "database"
	// QueryParamsKey is the key used to store the query parameters of current request in the context.
	QueryParamsKey ContextKey = "queryParams"
)

// DatabaseMiddleware is a middleware that injects the database connection into the context.
func MockDatabaseMiddleware(next http.Handler, db database.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session := db.NewSession()
		ctx = context.WithValue(ctx, DatabaseKey, session)
		rWithDatabase := r.WithContext(ctx)
		next.ServeHTTP(w, rWithDatabase)
	})
}
