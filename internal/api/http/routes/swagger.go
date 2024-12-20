package routes

import (
	"net/http"

	"github.com/mini-maxit/backend/internal/api/http/utils"
)

type SwaggerRoute struct{}

func NewSwaggerRoute() SwaggerRoute {
	return SwaggerRoute{}
}

func (sw *SwaggerRoute) Docs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ReturnError(w, http.StatusNotFound, "Not found")
		return
	}

	utils.ReturnError(w, http.StatusNotImplemented, "Not implemented")
	// idk what is going on the url in the index.html is not working and not getting the api sepecs correctly
	// http.ServeFile(w, r, "docs/index.html")
}
