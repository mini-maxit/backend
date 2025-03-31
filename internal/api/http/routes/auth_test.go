package routes_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/api/http/routes"
	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestLogin(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	us := mock_service.NewMockUserService(ctrl)
	as := testutils.NewMockAuthService()
	route := routes.NewAuthRoute(us, as)
	db := &testutils.MockDatabase{}
	handler := testutils.MockDatabaseMiddleware(http.HandlerFunc(route.Login), db)
	server := httptest.NewServer(handler)
	defer server.Close()

	hashPass, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	correctUser := &models.User{
		Email:        "email@email.com",
		PasswordHash: string(hashPass),
	}
	as.SetUser(correctUser)

	t.Run("Accept only post", func(t *testing.T) {
		methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			assert.HTTPStatusCode(t, route.Login, method, "", nil, http.StatusMethodNotAllowed)
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
		db.Vaildate()
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

	t.Run("Success", func(t *testing.T) {
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
		response := &httputils.APIResponse[schemas.Session]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.IsType(t, schemas.Session{}, response.Data)
	})
}

func TestRegister(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	us := mock_service.NewMockUserService(ctrl)
	as := testutils.NewMockAuthService()
	route := routes.NewAuthRoute(us, as)
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

	existingUser := &models.User{
		Email: "existing@email.com",
	}

	as.SetUser(existingUser)

	t.Run("Accept only post", func(t *testing.T) {
		methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			assert.HTTPStatusCode(t, route.Register, method, "", nil, http.StatusMethodNotAllowed)
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
				ConfirmPassword string `json:"confirm_password"`
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
				ConfirmPassword string `json:"confirm_password"`
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
			testInvalidRequestBody(t, server.URL, body)
		}
	})

	t.Run("Transaction was not started by middleware", func(t *testing.T) {
		testTxNotStarted(t, server.URL, db, correctRequest)
	})

	t.Run("User already exists", func(t *testing.T) {
		email := correctRequest.Email
		correctRequest.Email = existingUser.Email
		testUserAlreadyExists(t, server.URL, correctRequest)
		correctRequest.Email = email
	})

	t.Run("Success", func(t *testing.T) {
		testSuccess(t, server.URL, correctRequest)
	})
}

func testInvalidRequestBody(t *testing.T, url string, body any) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
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

func testTxNotStarted(t *testing.T, url string, db *testutils.MockDatabase, request schemas.UserRegisterRequest) {
	jsonBody, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}
	db.Invalidate()
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	db.Vaildate()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	bodyString := string(bodyBytes)

	assert.Contains(t, bodyString, "Transaction was not started by middleware")
}

func testUserAlreadyExists(t *testing.T, url string, request schemas.UserRegisterRequest) {
	jsonBody, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
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
}

func testSuccess(t *testing.T, url string, request schemas.UserRegisterRequest) {
	jsonBody, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	response := &httputils.APIResponse[schemas.Session]{}
	err = json.Unmarshal(bodyBytes, response)
	require.NoError(t, err)

	assert.IsType(t, schemas.Session{}, response.Data)
}
