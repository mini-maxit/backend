package routes

import (
	"net/http"

	"github.com/mini-maxit/backend/internal/api/http/utils"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/mini-maxit/backend/docs"
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

	httpSwagger.Handler(httpSwagger.URL("doc.json")).ServeHTTP(w, r)
}
