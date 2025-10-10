package routes_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/api/http/responses"
	"github.com/mini-maxit/backend/internal/api/http/routes"
	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/schemas"
	myerrors "github.com/mini-maxit/backend/package/errors"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

const refreshTokenCookieName = "refresh_token"

func TestLogin(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	us := mock_service.NewMockUserService(ctrl)
	as := mock_service.NewMockAuthService(ctrl)
	route := routes.NewAuthRoute(us, as, "/auth/refresh")
	db := &testutils.MockDatabase{}
	handler := testutils.MockDatabaseMiddleware(http.HandlerFunc(route.Login), db)
	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("Accept only post", func(t *testing.T) {
		methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Invalid request body", func(t *testing.T) {
		tt := []struct {
			body any
			msg  string
		}{
			{
				body: struct {
					Email string `json:"email"`
				}{
					Email: "email",
				},
				msg: "Invalid request body",
			},
			{
				body: struct {
					Email    string `json:"email"`
					Password string `json:"password"`
					Invalid  string `json:"invalid"`
				}{
					Email:    "email@email.com",
					Password: "password",
					Invalid:  "invalid",
				},
				msg: "Invalid request body",
			},
		}

		for _, tc := range tt {
			t.Logf("Testing with body: %v", tc.body)
			jsonBody, err := json.Marshal(tc.body)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonBody))
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}
			bodyString := string(bodyBytes)

			assert.Contains(t, bodyString, tc.msg)
		}
	})

	t.Run("Transaction was not started by middleware", func(t *testing.T) {
		body := struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}{
			Email:    "email@email.com",
			Password: "password",
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		db.Invalidate()
		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		db.Validate()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		bodyString := string(bodyBytes)

		assert.Contains(t, bodyString, "Transaction was not started by middleware")
	})

	t.Run("User not found", func(t *testing.T) {
		reqBody := struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}{
			Email:    "invalid@email.com",
			Password: "password",
		}
		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		as.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, myerrors.ErrUserNotFound).Times(1)

		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		bodyString := string(bodyBytes)

		assert.Contains(t, bodyString, "User not found")
	})

	t.Run("Invalid credentials", func(t *testing.T) {
		reqBody := struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}{
			Email:    "email@email.com",
			Password: "invalid",
		}
		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		as.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, myerrors.ErrInvalidCredentials).Times(1)

		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		bodyString := string(bodyBytes)

		assert.Contains(t, bodyString, "Invalid credentials")
	})

	t.Run("Internal server error", func(t *testing.T) {
		body := struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}{
			Email:    "email@email.com",
			Password: "password",
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		as.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, gorm.ErrInvalidDB).Times(1)

		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		bodyString := string(bodyBytes)

		assert.Contains(t, bodyString, "Internal Server Error")
	})

	t.Run("Success", func(t *testing.T) {
		body := struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}{
			Email:    "test@email.com",
			Password: "password",
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		expectedTokens := &schemas.JWTTokens{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
		}

		as.EXPECT().Login(gomock.Any(), gomock.Any()).Return(expectedTokens, nil).Times(1)

		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		response := &httputils.APIResponse[responses.AuthResponse]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.IsType(t, responses.AuthResponse{}, response.Data)

		// Check that refresh token cookie is set
		cookies := resp.Cookies()
		var refreshTokenCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == refreshTokenCookieName {
				refreshTokenCookie = cookie
				break
			}
		}
		assert.NotNil(t, refreshTokenCookie)
		assert.Equal(t, expectedTokens.RefreshToken, refreshTokenCookie.Value)
	})
}

func TestRegister(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	us := mock_service.NewMockUserService(ctrl)
	as := mock_service.NewMockAuthService(ctrl)
	route := routes.NewAuthRoute(us, as, "/auth/refresh")
	db := &testutils.MockDatabase{}
	handler := testutils.MockDatabaseMiddleware(http.HandlerFunc(route.Register), db)
	server := httptest.NewServer(handler)
	defer server.Close()

	correctRequest := schemas.UserRegisterRequest{
		Name:            "name",
		Surname:         "surname",
		Email:           "email@email.com",
		Username:        "username",
		Password:        "HardPassowrd123!",
		ConfirmPassword: "HardPassowrd123!",
	}

	t.Run("Accept only post", func(t *testing.T) {
		methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Invalid request body", func(t *testing.T) {
		invalidBodies := []any{
			struct {
				Email string `json:"email"`
			}{
				Email: "email",
			},
			struct {
				Name            string `json:"name"`
				Surname         string `json:"surname"`
				Username        string `json:"username"`
				Email           string `json:"email"`
				Password        string `json:"password"`
				ConfirmPassword string `json:"confirmPassword"`
				Invalid         string `json:"invalid"`
			}{
				Name:            "name",
				Surname:         "surname",
				Username:        "username",
				Email:           "email",
				Password:        "HardPassowrd123!",
				ConfirmPassword: "HardPassowrd123!",
				Invalid:         "invalid",
			},
			struct {
				Name            string `json:"name"`
				Surname         string `json:"surname"`
				Username        string `json:"username"`
				Email           string `json:"email"`
				Password        string `json:"password"`
				ConfirmPassword string `json:"confirmPassword"`
			}{
				Name:            "name",
				Surname:         "surname",
				Username:        "username",
				Email:           "email",
				Password:        "HardPassowrd123!",
				ConfirmPassword: "HardPassowrd123!",
			},
		}

		for _, body := range invalidBodies {
			jsonBody, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonBody))
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}
			bodyString := string(bodyBytes)

			assert.Contains(t, bodyString, "Invalid request body")
		}
	})

	t.Run("Transaction was not started by middleware", func(t *testing.T) {
		jsonBody, err := json.Marshal(correctRequest)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		db.Invalidate()
		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		db.Validate()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		bodyString := string(bodyBytes)

		assert.Contains(t, bodyString, "Transaction was not started by middleware")
	})

	t.Run("User already exists", func(t *testing.T) {
		as.EXPECT().Register(gomock.Any(), gomock.Any()).Return(nil, myerrors.ErrUserAlreadyExists).Times(1)

		jsonBody, err := json.Marshal(correctRequest)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		bodyString := string(bodyBytes)

		assert.Contains(t, bodyString, "user already exists")
	})

	t.Run("Internal server error", func(t *testing.T) {
		as.EXPECT().Register(gomock.Any(), gomock.Any()).Return(nil, gorm.ErrInvalidDB).Times(1)

		jsonBody, err := json.Marshal(correctRequest)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		bodyString := string(bodyBytes)

		assert.Contains(t, bodyString, "Internal Server Error")
	})

	t.Run("Success", func(t *testing.T) {
		expectedTokens := &schemas.JWTTokens{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
		}

		as.EXPECT().Register(gomock.Any(), gomock.Any()).Return(expectedTokens, nil).Times(1)

		jsonBody, err := json.Marshal(correctRequest)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		response := &httputils.APIResponse[responses.AuthResponse]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)

		assert.IsType(t, responses.AuthResponse{}, response.Data)

		// Check that refresh token cookie is set
		cookies := resp.Cookies()
		var refreshTokenCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == refreshTokenCookieName {
				refreshTokenCookie = cookie
				break
			}
		}
		assert.NotNil(t, refreshTokenCookie)
		assert.Equal(t, "refresh_token", refreshTokenCookie.Value)
	})
}

func TestRefreshToken(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	us := mock_service.NewMockUserService(ctrl)
	as := mock_service.NewMockAuthService(ctrl)
	route := routes.NewAuthRoute(us, as, "/auth/refresh")
	db := &testutils.MockDatabase{}
	handler := testutils.MockDatabaseMiddleware(http.HandlerFunc(route.RefreshToken), db)
	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("Accept only post", func(t *testing.T) {
		methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Missing refresh token cookie", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, server.URL, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		bodyString := string(bodyBytes)
		assert.Contains(t, bodyString, "Refresh token cookie not found")
	})

	t.Run("Transaction was not started by middleware", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, server.URL, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Add refresh token cookie
		req.AddCookie(&http.Cookie{
			Name:  refreshTokenCookieName,
			Value: "valid_refresh_token",
		})

		db.Invalidate()
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		db.Validate()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		bodyString := string(bodyBytes)

		assert.Contains(t, bodyString, "Transaction was not started by middleware")
	})

	t.Run("Invalid or expired refresh token", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, server.URL, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Add refresh token cookie
		req.AddCookie(&http.Cookie{
			Name:  "refresh_token",
			Value: "invalid_refresh_token",
		})

		as.EXPECT().RefreshTokens(gomock.Any(), gomock.Any()).Return(nil, myerrors.ErrInvalidToken).Times(1)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		bodyString := string(bodyBytes)

		assert.Contains(t, bodyString, "Invalid or expired refresh token")
	})

	t.Run("Internal server error", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, server.URL, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Add refresh token cookie
		req.AddCookie(&http.Cookie{
			Name:  "refresh_token",
			Value: "valid_refresh_token",
		})

		as.EXPECT().RefreshTokens(gomock.Any(), gomock.Any()).Return(nil, gorm.ErrInvalidDB).Times(1)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		bodyString := string(bodyBytes)

		assert.Contains(t, bodyString, "Internal Server Error")
	})

	t.Run("Success", func(t *testing.T) {
		expectedTokens := &schemas.JWTTokens{
			AccessToken:  "new_access_token",
			RefreshToken: "new_refresh_token",
		}

		as.EXPECT().RefreshTokens(gomock.Any(), gomock.Any()).Return(expectedTokens, nil).Times(1)

		req, err := http.NewRequest(http.MethodPost, server.URL, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Add refresh token cookie
		req.AddCookie(&http.Cookie{
			Name:  "refresh_token",
			Value: "valid_refresh_token",
		})

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		response := &httputils.APIResponse[responses.AuthResponse]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)

		assert.IsType(t, responses.AuthResponse{}, response.Data)
		assert.Equal(t, "new_access_token", response.Data.AccessToken)

		// Check that new refresh token cookie is set
		cookies := resp.Cookies()
		var refreshTokenCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "refresh_token" {
				refreshTokenCookie = cookie
				break
			}
		}
		assert.NotNil(t, refreshTokenCookie)
		assert.Equal(t, "new_refresh_token", refreshTokenCookie.Value)
	})
}
