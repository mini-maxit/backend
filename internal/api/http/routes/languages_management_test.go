package routes_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/api/http/routes"
	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/schemas"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLanguagesManagement_GetAllLanguages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ls := mock_service.NewMockLanguageService(ctrl)
	route := routes.NewLanguagesManagementRoute(ls)
	db := &testutils.MockDatabase{}
	handler := httputils.MockDatabaseMiddleware(http.HandlerFunc(route.GetAllLanguages), db)
	server := httptest.NewServer(handler)
	defer server.Close()

	expected := []schemas.LanguageConfig{
		{ID: 1, Type: "python", Version: "3.10", FileExtension: ".py", IsDisabled: false},
		{ID: 2, Type: "CPP", Version: "20", FileExtension: ".cpp", IsDisabled: true},
	}
	ls.EXPECT().GetAll(gomock.Any()).Return(expected, nil).Times(1)

	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body httputils.APIResponse[[]schemas.LanguageConfig]
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	assert.Len(t, body.Data, 2)
	assert.True(t, body.Data[1].IsDisabled)
}

func TestLanguagesManagement_GetAllLanguages_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ls := mock_service.NewMockLanguageService(ctrl)
	route := routes.NewLanguagesManagementRoute(ls)
	db := &testutils.MockDatabase{}
	handler := httputils.MockDatabaseMiddleware(http.HandlerFunc(route.GetAllLanguages), db)
	server := httptest.NewServer(handler)
	defer server.Close()

	ls.EXPECT().GetAll(gomock.Any()).Return(nil, assert.AnError).Times(1)

	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestLanguagesManagement_ToggleLanguageVisibility(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ls := mock_service.NewMockLanguageService(ctrl)
	route := routes.NewLanguagesManagementRoute(ls)
	db := &testutils.MockDatabase{}
	handler := httputils.MockDatabaseMiddleware(http.HandlerFunc(route.ToggleLanguageVisibility), db)

	t.Run("Toggle valid language", func(t *testing.T) {
		ls.EXPECT().ToggleLanguageVisibility(gomock.Any(), int64(3)).Return(nil).Times(1)

		req := httptest.NewRequest(http.MethodPatch, "/languages/3", bytes.NewBufferString(`{}`))
		req = mux.SetURLVars(req, map[string]string{"id": "3"})
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid language ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/languages/abc", bytes.NewBufferString(`{}`))
		req = mux.SetURLVars(req, map[string]string{"id": "abc"})
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Missing language ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/languages/", bytes.NewBufferString(`{}`))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
