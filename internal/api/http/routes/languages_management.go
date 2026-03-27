package routes

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
)

type LanguagesManagementRoute interface {
	GetAllLanguages(w http.ResponseWriter, r *http.Request)
	ToggleLanguageVisibility(w http.ResponseWriter, r *http.Request)
}

type languagesManagementRoute struct {
	languageService service.LanguageService
	logger          *zap.SugaredLogger
}

// GetAllLanguages godoc
//
//	@Tags			languages-management
//	@Summary		Get all languages
//	@Description	Get all language configurations
//	@Produce		json
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[[]schemas.LanguageConfig]
//	@Router			/languages-management/languages [get]
func (lr *languagesManagementRoute) GetAllLanguages(w http.ResponseWriter, r *http.Request) {
	db := httputils.GetDatabase(r)

	languages, err := lr.languageService.GetAll(db)
	if err != nil {
		httputils.HandleServiceError(w, err, db, lr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, languages)
}

// ToggleLanguageVisibility godoc
//
//	@Tags			languages-management
//	@Summary		Toggle language visibility
//	@Description	Toggle the visibility (enabled/disabled) state of a language
//	@Produce		json
//	@Param			id	path		int	true	"Language ID"
//	@Failure		400	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/languages-management/languages/{id} [patch]
func (lr *languagesManagementRoute) ToggleLanguageVisibility(w http.ResponseWriter, r *http.Request) {
	languageIDStr := httputils.GetPathValue(r, "id")
	if languageIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Language ID is required.")
		return
	}

	languageID, err := strconv.ParseInt(languageIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid language ID.")
		return
	}

	db := httputils.GetDatabase(r)

	err = lr.languageService.ToggleLanguageVisibility(db, languageID)
	if err != nil {
		httputils.HandleServiceError(w, err, db, lr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Language visibility toggled successfully"))
}

func NewLanguagesManagementRoute(languageService service.LanguageService) LanguagesManagementRoute {
	route := &languagesManagementRoute{
		languageService: languageService,
		logger:          utils.NewNamedLogger("languages-management-route"),
	}

	if err := utils.ValidateStruct(*route); err != nil {
		panic(err)
	}
	return route
}

func RegisterLanguagesManagementRoutes(mux *mux.Router, route LanguagesManagementRoute) {
	mux.HandleFunc("/languages", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetAllLanguages(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
	mux.HandleFunc("/languages/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPatch:
			route.ToggleLanguageVisibility(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
}
