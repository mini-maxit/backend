package routes

import (
	"net/http"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
)

type SwaggerRoute struct{}

func NewSwaggerRoute() SwaggerRoute {
	return SwaggerRoute{}
}

func (sw *SwaggerRoute) Docs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusNotFound, "Not found")
		return
	}

	http.ServeFile(w, r, "docs/index.html")
}
