package middleware

import (
	"context"
	"net/http"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/database"
)

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
		session := db.NewSession()
		tx, err := session.BeginTransaction()
		if err != nil {
			httputils.ReturnError(w, http.StatusInternalServerError, "Failed to start transaction. "+err.Error())
			return
		}
		ctx = context.WithValue(ctx, httputils.DatabaseKey, session)
		rWithDatabase := r.WithContext(ctx)
		wrappedWriter := &ResponseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
		defer func() {
			if session.ShouldRollback() {
				tx.Rollback()
			} else {
				err := session.Commit()
				if err != nil {
					tx.Rollback()
					httputils.ReturnError(w, http.StatusInternalServerError, "Failed to commit transaction. "+err.Error())
					return
				}
			}
			session = nil
		}()
		next.ServeHTTP(wrappedWriter, rWithDatabase)
	})
}
