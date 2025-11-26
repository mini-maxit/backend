package config_test

import (
	"os"
	"testing"

	"github.com/mini-maxit/backend/internal/config"
	"github.com/stretchr/testify/require"
)

// All environment variables used by NewConfig.
var allEnvVars = []string{
	"DB_HOST",
	"DB_PORT",
	"DB_USER",
	"DB_PASSWORD",
	"DB_NAME",
	"APP_PORT",
	"API_REFRESH_TOKEN_PATH",
	"FILE_STORAGE_HOST",
	"FILE_STORAGE_PORT",
	"QUEUE_NAME",
	"RESPONSE_QUEUE_NAME",
	"QUEUE_HOST",
	"QUEUE_PORT",
	"QUEUE_USER",
	"QUEUE_PASSWORD",
	"JWT_SECRET_KEY",
	"DUMP",
}

// Base valid environment values.
var baseEnv = map[string]string{
	"DB_HOST":                "localhost",
	"DB_PORT":                "5432",
	"DB_USER":                "user",
	"DB_PASSWORD":            "pass",
	"DB_NAME":                "appdb",
	"APP_PORT":               "9090",
	"API_REFRESH_TOKEN_PATH": "/api/v1/auth/refresh-custom",
	"FILE_STORAGE_HOST":      "filesvc",
	"FILE_STORAGE_PORT":      "9100",
	"QUEUE_NAME":             "custom_worker_queue",
	"RESPONSE_QUEUE_NAME":    "custom_worker_response_queue",
	"QUEUE_HOST":             "queuehost",
	"QUEUE_PORT":             "5673",
	"QUEUE_USER":             "queueuser",
	"QUEUE_PASSWORD":         "queuepass",
	"JWT_SECRET_KEY":         "supersecret",
	"DUMP":                   "true",
}

func unsetAll() {
	for _, k := range allEnvVars {
		_ = os.Unsetenv(k)
	}
}

func setEnv(values map[string]string) {
	for k, v := range values {
		_ = os.Setenv(k, v)
	}
}

func TestNewConfig_SuccessFullEnv(t *testing.T) {
	unsetAll()
	setEnv(baseEnv)

	cfg := config.NewConfig()
	require.NotNil(t, cfg)

	// DB
	require.Equal(t, baseEnv["DB_HOST"], cfg.DB.Host)
	require.Equal(t, uint16(5432), cfg.DB.Port)
	require.Equal(t, baseEnv["DB_USER"], cfg.DB.User)
	require.Equal(t, baseEnv["DB_PASSWORD"], cfg.DB.Password)
	require.Equal(t, baseEnv["DB_NAME"], cfg.DB.Name)

	// API
	require.Equal(t, uint16(9090), cfg.API.Port)
	require.Equal(t, baseEnv["API_REFRESH_TOKEN_PATH"], cfg.API.RefreshTokenPath)

	// Broker
	require.Equal(t, baseEnv["QUEUE_NAME"], cfg.Broker.QueueName)
	require.Equal(t, baseEnv["RESPONSE_QUEUE_NAME"], cfg.Broker.ResponseQueueName)
	require.Equal(t, baseEnv["QUEUE_HOST"], cfg.Broker.Host)
	require.Equal(t, uint16(5673), cfg.Broker.Port)
	require.Equal(t, baseEnv["QUEUE_USER"], cfg.Broker.User)
	require.Equal(t, baseEnv["QUEUE_PASSWORD"], cfg.Broker.Password)

	// File storage URL composition
	require.Equal(t, "http://"+baseEnv["FILE_STORAGE_HOST"]+":"+baseEnv["FILE_STORAGE_PORT"], cfg.FileStorageURL)

	// JWT
	require.Equal(t, baseEnv["JWT_SECRET_KEY"], cfg.JWTSecretKey)

	// Dump flag
	require.True(t, cfg.Dump)
}

func TestNewConfig_DefaultsAndOptionalMissing(t *testing.T) {
	unsetAll()

	// Required variables only (omit optional ones)
	minimal := map[string]string{
		"DB_PORT":           "5432",
		"DB_USER":           "user",
		"DB_NAME":           "db",
		"FILE_STORAGE_HOST": "fs",
		"FILE_STORAGE_PORT": "9000",
		"QUEUE_HOST":        "qhost",
		"QUEUE_PORT":        "5672",
		"QUEUE_USER":        "quser",
		"QUEUE_PASSWORD":    "qpass",
		"JWT_SECRET_KEY":    "secret",
		// Omit APP_PORT, API_REFRESH_TOKEN_PATH, QUEUE_NAME, RESPONSE_QUEUE_NAME, DB_HOST, DB_PASSWORD, DUMP
	}
	setEnv(minimal)

	cfg := config.NewConfig()
	require.NotNil(t, cfg)

	// Defaults applied
	require.Equal(t, uint16(8080), cfg.API.Port)
	require.Equal(t, "/api/v1/auth/refresh", cfg.API.RefreshTokenPath)
	require.Equal(t, "worker_queue", cfg.Broker.QueueName)
	require.Equal(t, "worker_response_queue", cfg.Broker.ResponseQueueName)

	// Optional missing values preserved as empty string
	require.Empty(t, cfg.DB.Host)     // Warns but does not set "localhost"
	require.Empty(t, cfg.DB.Password) // Warns but stays empty

	// Dump default false
	require.False(t, cfg.Dump)
}

func TestNewConfig_DumpFlagFalseWhenMissing(t *testing.T) {
	unsetAll()
	env := map[string]string{
		"DB_PORT":           "5432",
		"DB_USER":           "user",
		"DB_NAME":           "db",
		"FILE_STORAGE_HOST": "fs",
		"FILE_STORAGE_PORT": "9000",
		"QUEUE_HOST":        "qhost",
		"QUEUE_PORT":        "5672",
		"QUEUE_USER":        "quser",
		"QUEUE_PASSWORD":    "qpass",
		"JWT_SECRET_KEY":    "secret",
	}
	setEnv(env)

	cfg := config.NewConfig()
	require.NotNil(t, cfg)
	require.False(t, cfg.Dump)
}

func TestNewConfig_PanicsWhenRequiredMissing(t *testing.T) {
	requiredMissing := []string{
		"DB_PORT",
		"DB_USER",
		"DB_NAME",
		"FILE_STORAGE_HOST",
		"FILE_STORAGE_PORT",
		"QUEUE_HOST",
		"QUEUE_PORT",
		"QUEUE_USER",
		"QUEUE_PASSWORD",
		"JWT_SECRET_KEY",
	}

	for _, missing := range requiredMissing {
		t.Run("missing_"+missing, func(t *testing.T) {
			unsetAll()
			setEnv(baseEnv) // start from full valid
			_ = os.Unsetenv(missing)

			didPanic := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						didPanic = true
					}
				}()
				_ = config.NewConfig()
			}()
			require.True(t, didPanic, "expected panic when %s is missing", missing)
		})
	}
}

func TestNewConfig_PanicsOnInvalidPortValues(t *testing.T) {
	cases := []struct {
		name string
		vars map[string]string
	}{
		{
			name: "invalid DB_PORT",
			vars: map[string]string{"DB_PORT": "notint"},
		},
		{
			name: "invalid APP_PORT",
			vars: map[string]string{"APP_PORT": "invalid"},
		},
		{
			name: "invalid FILE_STORAGE_PORT",
			vars: map[string]string{"FILE_STORAGE_PORT": "bad"},
		},
		{
			name: "invalid QUEUE_PORT",
			vars: map[string]string{"QUEUE_PORT": "oops"},
		},
		{
			name: "out_of_range DB_PORT",
			vars: map[string]string{"DB_PORT": "70000"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			unsetAll()
			setEnv(baseEnv)
			for k, v := range tc.vars {
				t.Setenv(k, v)
			}

			didPanic := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						didPanic = true
					}
				}()
				_ = config.NewConfig()
			}()
			require.True(t, didPanic, "expected panic for case %s", tc.name)
		})
	}
}
