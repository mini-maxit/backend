//nolint:testpackage // testing private default values
package config

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

const baseDBUser = "user"

// All environment variables used by NewConfig.
var allEnvVars = []string{
	envDBHost,
	envDBPort,
	envDBUser,
	envDBPassword,
	envDBName,
	envAPPPort,
	envRefreshToken,
	envFileStorageHost,
	envFileStoragePort,
	envQueueName,
	envResponseQueue,
	envQueueHost,
	envQueuePort,
	envQueueUser,
	envQueuePassword,
	envJWTSecretKey,
	envDump,
}

// Base valid environment values.
var baseEnv = map[string]string{
	envDBHost:          "localhost",
	envDBPort:          "5432",
	envDBUser:          baseDBUser,
	envDBPassword:      "pass",
	envDBName:          "appdb",
	envAPPPort:         "9090",
	envRefreshToken:    "/api/v1/auth/refresh-custom",
	envFileStorageHost: "filesvc",
	envFileStoragePort: "9100",
	envQueueName:       "custom_worker_queue",
	envResponseQueue:   "custom_worker_response_queue",
	envQueueHost:       "queuehost",
	envQueuePort:       "5673",
	envQueueUser:       "queueuser",
	envQueuePassword:   "queuepass",
	envJWTSecretKey:    "supersecret",
	envDump:            "true",
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

	cfg := NewConfig()
	require.NotNil(t, cfg)

	// DB
	require.Equal(t, baseEnv[envDBHost], cfg.DB.Host)
	require.Equal(t, uint16(5432), cfg.DB.Port)
	require.Equal(t, baseEnv[envDBUser], cfg.DB.User)
	require.Equal(t, baseEnv[envDBPassword], cfg.DB.Password)
	require.Equal(t, baseEnv[envDBName], cfg.DB.Name)

	// API
	require.Equal(t, uint16(9090), cfg.API.Port)
	require.Equal(t, baseEnv[envRefreshToken], cfg.API.RefreshTokenPath)

	// Broker
	require.Equal(t, baseEnv[envQueueName], cfg.Broker.QueueName)
	require.Equal(t, baseEnv[envResponseQueue], cfg.Broker.ResponseQueueName)
	require.Equal(t, baseEnv[envQueueHost], cfg.Broker.Host)
	require.Equal(t, uint16(5673), cfg.Broker.Port)
	require.Equal(t, baseEnv[envQueueUser], cfg.Broker.User)
	require.Equal(t, baseEnv[envQueuePassword], cfg.Broker.Password)

	// File storage URL composition
	require.Equal(t, "http://"+baseEnv[envFileStorageHost]+":"+baseEnv[envFileStoragePort], cfg.FileStorageURL)

	// JWT
	require.Equal(t, baseEnv[envJWTSecretKey], cfg.JWTSecretKey)

	// Dump flag
	require.True(t, cfg.Dump)
}

func TestNewConfig_DefaultsAndOptionalMissing(t *testing.T) {
	unsetAll()

	// Required variables only (omit optional ones)
	minimal := map[string]string{
		envDBPort:          "5432",
		envDBUser:          baseDBUser,
		envDBName:          "db",
		envFileStorageHost: "fs",
		envFileStoragePort: "9000",
		envQueueHost:       "qhost",
		envQueuePort:       "5672",
		envQueueUser:       "quser",
		envQueuePassword:   "qpass",
		envJWTSecretKey:    "secret",
		// Omit APP_PORT, API_REFRESH_TOKEN_PATH, QUEUE_NAME, RESPONSE_QUEUE_NAME, DB_HOST, DB_PASSWORD, DUMP
	}
	setEnv(minimal)

	cfg := NewConfig()
	require.NotNil(t, cfg)

	// Defaults applied
	port, _ := strconv.ParseUint(defaultAPIPort, 10, 16)
	require.Equal(t, uint16(port), cfg.API.Port)
	require.Equal(t, defaultAPIRefreshTokenPath, cfg.API.RefreshTokenPath)
	require.Equal(t, defaultQueueName, cfg.Broker.QueueName)
	require.Equal(t, defaultResponseQueueName, cfg.Broker.ResponseQueueName)

	// Optional missing values preserved as empty string
	require.Empty(t, cfg.DB.Host)     // Warns but does not set "localhost"
	require.Empty(t, cfg.DB.Password) // Warns but stays empty

	// Dump default false
	require.False(t, cfg.Dump)
}

func TestNewConfig_DumpFlagFalseWhenMissing(t *testing.T) {
	unsetAll()
	env := map[string]string{
		envDBPort:          "5432",
		envDBUser:          baseDBUser,
		envDBName:          "db",
		envFileStorageHost: "fs",
		envFileStoragePort: "9000",
		envQueueHost:       "qhost",
		envQueuePort:       "5672",
		envQueueUser:       "quser",
		envQueuePassword:   "qpass",
		envJWTSecretKey:    "secret",
	}
	setEnv(env)

	cfg := NewConfig()
	require.NotNil(t, cfg)
	require.False(t, cfg.Dump)
}

func TestNewConfig_PanicsWhenRequiredMissing(t *testing.T) {
	requiredMissing := []string{
		envDBPort,
		envDBUser,
		envDBName,
		envFileStorageHost,
		envFileStoragePort,
		envQueueHost,
		envQueuePort,
		envQueueUser,
		envQueuePassword,
		envJWTSecretKey,
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
				_ = NewConfig()
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
			vars: map[string]string{envDBPort: "notint"},
		},
		{
			name: "invalid APP_PORT",
			vars: map[string]string{envAPPPort: "invalid"},
		},
		{
			name: "invalid FILE_STORAGE_PORT",
			vars: map[string]string{envFileStoragePort: "bad"},
		},
		{
			name: "invalid QUEUE_PORT",
			vars: map[string]string{envQueuePort: "oops"},
		},
		{
			name: "out_of_range DB_PORT",
			vars: map[string]string{envDBPort: "70000"},
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
				_ = NewConfig()
			}()
			require.True(t, didPanic, "expected panic for case %s", tc.name)
		})
	}
}
