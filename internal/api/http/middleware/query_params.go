package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
)

func QueryParamsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		queryParams, err := httputils.GetQueryParams(&query)
		if err != nil {
			httputils.ReturnError(w, http.StatusBadRequest, err.Error())
			return
		}

		log.Printf("Query params: %+v", queryParams)
		ctx := context.WithValue(r.Context(), QueryParamsKey, queryParams)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
