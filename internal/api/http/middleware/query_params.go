package middleware

import (
	"context"
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

		ctx := context.WithValue(r.Context(), httputils.QueryParamsKey, queryParams)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
