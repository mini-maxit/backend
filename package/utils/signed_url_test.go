package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testBaseURL = "http://example.com/file.pdf"

func TestSignedURLGenerator_GenerateSignedURL(t *testing.T) {
	generator := NewSignedURLGenerator("test-secret-key", 300)

	signedURL, err := generator.GenerateSignedURL(testBaseURL)
	require.NoError(t, err)
	assert.NotEmpty(t, signedURL)

	// Parse the signed URL and check for required parameters
	parsedURL, err := url.Parse(signedURL)
	require.NoError(t, err)

	queryParams := parsedURL.Query()
	assert.NotEmpty(t, queryParams.Get("expires"), "expires parameter should be present")
	assert.NotEmpty(t, queryParams.Get("signature"), "signature parameter should be present")

	// Verify the expiration is in the future
	expiresStr := queryParams.Get("expires")
	expiresAt, err := strconv.ParseInt(expiresStr, 10, 64)
	require.NoError(t, err)
	assert.Greater(t, expiresAt, time.Now().Unix(), "expiration should be in the future")
}

func TestSignedURLGenerator_GenerateSignedURL_WithExistingQueryParams(t *testing.T) {
	generator := NewSignedURLGenerator("test-secret-key", 300)
	baseURL := "http://example.com/file.pdf?existing=param&another=value"

	signedURL, err := generator.GenerateSignedURL(baseURL)
	require.NoError(t, err)

	parsedURL, err := url.Parse(signedURL)
	require.NoError(t, err)

	queryParams := parsedURL.Query()
	assert.Equal(t, "param", queryParams.Get("existing"), "existing params should be preserved")
	assert.Equal(t, "value", queryParams.Get("another"), "existing params should be preserved")
	assert.NotEmpty(t, queryParams.Get("expires"))
	assert.NotEmpty(t, queryParams.Get("signature"))
}

func TestSignedURLGenerator_VerifySignedURL_Valid(t *testing.T) {
	generator := NewSignedURLGenerator("test-secret-key", 300)

	signedURL, err := generator.GenerateSignedURL(testBaseURL)
	require.NoError(t, err)

	// Verify the signed URL
	err = generator.VerifySignedURL(signedURL)
	assert.NoError(t, err, "valid signed URL should verify successfully")
}

func TestSignedURLGenerator_VerifySignedURL_Expired(t *testing.T) {
	generator := NewSignedURLGenerator("test-secret-key", 300)

	// Manually create an expired signed URL
	parsedURL, err := url.Parse(testBaseURL)
	require.NoError(t, err)

	// Set expiration to 1 second in the past
	expiresAt := time.Now().Add(-1 * time.Second).Unix()
	queryParams := parsedURL.Query()
	queryParams.Set("expires", strconv.FormatInt(expiresAt, 10))
	parsedURL.RawQuery = queryParams.Encode()

	// Generate signature for the expired URL
	stringToSign := parsedURL.String()
	h := hmac.New(sha256.New, []byte("test-secret-key"))
	h.Write([]byte(stringToSign))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))

	queryParams.Set("signature", signature)
	parsedURL.RawQuery = queryParams.Encode()
	expiredURL := parsedURL.String()

	// Verify should fail due to expiration
	err = generator.VerifySignedURL(expiredURL)
	assert.ErrorIs(t, err, ErrSignedURLExpired, "expired URL should return ErrSignedURLExpired")
}

func TestSignedURLGenerator_VerifySignedURL_InvalidSignature(t *testing.T) {
	generator := NewSignedURLGenerator("test-secret-key", 300)

	signedURL, err := generator.GenerateSignedURL(testBaseURL)
	require.NoError(t, err)

	// Tamper with the URL by changing a query parameter
	parsedURL, err := url.Parse(signedURL)
	require.NoError(t, err)

	queryParams := parsedURL.Query()
	queryParams.Set("tampered", "true")
	parsedURL.RawQuery = queryParams.Encode()
	tamperedURL := parsedURL.String()

	// Verify should fail due to invalid signature
	err = generator.VerifySignedURL(tamperedURL)
	assert.ErrorIs(t, err, ErrSignedURLInvalidSignature, "tampered URL should return ErrSignedURLInvalidSignature")
}

func TestSignedURLGenerator_VerifySignedURL_WrongSecretKey(t *testing.T) {
	generator1 := NewSignedURLGenerator("secret-key-1", 300)
	generator2 := NewSignedURLGenerator("secret-key-2", 300)

	// Generate URL with generator1
	signedURL, err := generator1.GenerateSignedURL(testBaseURL)
	require.NoError(t, err)

	// Try to verify with generator2 (different secret)
	err = generator2.VerifySignedURL(signedURL)
	assert.ErrorIs(t, err, ErrSignedURLInvalidSignature, "URL signed with different key should fail verification")
}

func TestSignedURLGenerator_VerifySignedURL_MissingExpires(t *testing.T) {
	generator := NewSignedURLGenerator("test-secret-key", 300)
	invalidURL := "http://example.com/file.pdf?signature=abc123"

	err := generator.VerifySignedURL(invalidURL)
	assert.ErrorIs(t, err, ErrSignedURLMissingParams, "URL without expires should return ErrSignedURLMissingParams")
}

func TestSignedURLGenerator_VerifySignedURL_MissingSignature(t *testing.T) {
	generator := NewSignedURLGenerator("test-secret-key", 300)
	invalidURL := "http://example.com/file.pdf?expires=123456789"

	err := generator.VerifySignedURL(invalidURL)
	assert.ErrorIs(t, err, ErrSignedURLMissingParams, "URL without signature should return ErrSignedURLMissingParams")
}

func TestSignedURLGenerator_TTLDuration(t *testing.T) {
	ttlSeconds := uint16(120)
	generator := NewSignedURLGenerator("test-secret-key", ttlSeconds)

	signedURL, err := generator.GenerateSignedURL(testBaseURL)
	require.NoError(t, err)

	parsedURL, err := url.Parse(signedURL)
	require.NoError(t, err)

	queryParams := parsedURL.Query()
	expiresStr := queryParams.Get("expires")
	expiresAt, err := strconv.ParseInt(expiresStr, 10, 64)
	require.NoError(t, err)

	expectedExpiration := time.Now().Add(time.Duration(ttlSeconds) * time.Second).Unix()
	// Allow 2 second tolerance for test execution time
	assert.InDelta(t, expectedExpiration, expiresAt, 2, "expiration should match TTL")
}
