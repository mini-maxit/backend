package middleware

import (
	"context"
	"net/http"

	"github.com/mini-maxit/backend/internal/api/http/utils"
	"github.com/mini-maxit/backend/internal/database"
)

func requestSuccess(code int) bool {
	return code >= 200 && code < 300
}

type ResponseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *ResponseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *ResponseWriterWrapper) StatusCode() int {
	return w.statusCode
}

// DatabaseMiddleware is a middleware that injects the database connection into the context.
func DatabaseMiddleware(next http.Handler, db database.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tx := db.Connect().Begin()
		if tx.Error != nil {
			utils.ReturnError(w, http.StatusInternalServerError, "Failed to start transaction. "+tx.Error.Error())
			return
		}
		ctx = context.WithValue(ctx, DatabaseKey, db)
		rWithDatabase := r.WithContext(ctx)
		wrappedWriter := &ResponseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrappedWriter, rWithDatabase)
		if requestSuccess(wrappedWriter.StatusCode()) {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	})

}
