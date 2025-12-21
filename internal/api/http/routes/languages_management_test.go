package routes_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/api/http/routes"
	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/errors"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetAllLanguages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ls := mock_service.NewMockLanguageService(ctrl)
	route := routes.NewLanguagesManagementRoute(ls)
	db := &testutils.MockDatabase{}

	router := mux.NewRouter()
	router.HandleFunc("/languages", func(w http.ResponseWriter, r *http.Request) {
		route.GetAllLanguages(w, r)
	}).Methods(http.MethodGet)

	handler := httputils.MockDatabaseMiddleware(router, db)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	t.Run("Success", func(t *testing.T) {
		expectedLanguages := []schemas.LanguageConfig{
			{ID: 1, Type: "C", Version: "11", FileExtension: ".c", IsDisabled: false},
			{ID: 2, Type: "C++", Version: "17", FileExtension: ".cpp", IsDisabled: false},
		}

		ls.EXPECT().GetAll(gomock.Any()).Return(expectedLanguages, nil)

		req, err := http.NewRequest(http.MethodGet, server.URL+"/languages", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var apiResponse httputils.APIResponse[[]schemas.LanguageConfig]
		err = json.NewDecoder(resp.Body).Decode(&apiResponse)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		assert.Equal(t, expectedLanguages, apiResponse.Data)
	})

	t.Run("Service error", func(t *testing.T) {
		ls.EXPECT().GetAll(gomock.Any()).Return(nil, assert.AnError)

		req, err := http.NewRequest(http.MethodGet, server.URL+"/languages", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestToggleLanguageVisibility(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ls := mock_service.NewMockLanguageService(ctrl)
	route := routes.NewLanguagesManagementRoute(ls)
	db := &testutils.MockDatabase{}

	router := mux.NewRouter()
	router.HandleFunc("/languages/{id}", func(w http.ResponseWriter, r *http.Request) {
		route.ToggleLanguageVisibility(w, r)
	}).Methods(http.MethodPatch)

	handler := httputils.MockDatabaseMiddleware(router, db)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	t.Run("Invalid language ID", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPatch, server.URL+"/languages/invalid", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Language not found", func(t *testing.T) {
		ls.EXPECT().ToggleLanguageVisibility(gomock.Any(), int64(999)).Return(errors.ErrNotFound)

		req, err := http.NewRequest(http.MethodPatch, server.URL+"/languages/999", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		ls.EXPECT().ToggleLanguageVisibility(gomock.Any(), int64(1)).Return(nil)

		req, err := http.NewRequest(http.MethodPatch, server.URL+"/languages/1", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var apiResponse httputils.APIResponse[httputils.MessageResponse]
		err = json.NewDecoder(resp.Body).Decode(&apiResponse)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		assert.Equal(t, "Language visibility toggled successfully", apiResponse.Data.Message)
	})
}
