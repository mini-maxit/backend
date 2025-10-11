package routes

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestLogin(t *testing.T) {
	// Setup
	us := testutils.NewMockUserService()
	as := testutils.NewMockAuthService()
	route := NewAuthRoute(us, as)
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
			body interface{}
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
		response := &httputils.ApiResponse[schemas.Session]{}
		json.Unmarshal(bodyBytes, response)
		assert.IsType(t, schemas.Session{}, response.Data)
	})

}

func TestRegister(t *testing.T) {
	// Setup
	us := testutils.NewMockUserService()
	as := testutils.NewMockAuthService()
	route := NewAuthRoute(us, as)
	db := &testutils.MockDatabase{}
	handler := testutils.MockDatabaseMiddleware(http.HandlerFunc(route.Register), db)
	server := httptest.NewServer(handler)
	defer server.Close()
	correctRequest := schemas.UserRegisterRequest{
		Name:            "name",
		Surname:         "surname",
		Email:           "email@email.com",
		Username:        "username",
		Password:        "SuperStrongPassword1!",
		ConfirmPassword: "SuperStrongPassword1!",
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
		tt := []interface{}{
			struct {
				Email string `json:"email"`
			}{
				Email: "email",
			},
			struct {
				Name     string `json:"name"`
				Surname  string `json:"surname"`
				Username string `json:"username"`
				Email    string `json:"email"`
				Password string `json:"password"`
				Invalid  string `json:"invalid"`
			}{
				Name:     "name",
				Surname:  "surname",
				Username: "username",
				Email:    "email",
				Password: "password",
				Invalid:  "invalid",
			},
			struct {
				Name     string `json:"name"`
				Surname  string `json:"surname"`
				Username string `json:"username"`
				Email    string `json:"email"`
				Password string `json:"password"`
			}{
				Name:     "name",
				Surname:  "surname",
				Username: "username",
				Email:    "email",
				Password: "password",
			},
		}
		for _, body := range tt {
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

	t.Run("User already exists", func(t *testing.T) {
		email := correctRequest.Email
		correctRequest.Email = existingUser.Email
		jsonBody, err := json.Marshal(correctRequest)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		correctRequest.Email = email
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

	t.Run("Success", func(t *testing.T) {
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
		response := &httputils.ApiResponse[schemas.Session]{}
		json.Unmarshal(bodyBytes, response)
		assert.IsType(t, schemas.Session{}, response.Data)
	})

}
