package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

var (
	// ErrSignedURLExpired is returned when the signed URL has expired
	ErrSignedURLExpired = errors.New("signed URL has expired")
	// ErrSignedURLInvalidSignature is returned when the signature validation fails
	ErrSignedURLInvalidSignature = errors.New("signed URL has invalid signature")
	// ErrSignedURLMissingParams is returned when required parameters are missing
	ErrSignedURLMissingParams = errors.New("signed URL is missing required parameters")
)

// SignedURLGenerator generates and validates signed URLs
type SignedURLGenerator struct {
	secretKey []byte
	ttl       time.Duration
}

// NewSignedURLGenerator creates a new SignedURLGenerator
func NewSignedURLGenerator(secretKey string, ttlSeconds uint16) *SignedURLGenerator {
	return &SignedURLGenerator{
		secretKey: []byte(secretKey),
		ttl:       time.Duration(ttlSeconds) * time.Second,
	}
}

// GetSecretKey returns the secret key (for internal use)
func (g *SignedURLGenerator) GetSecretKey() []byte {
	return g.secretKey
}

// GenerateSignedURL generates a signed URL with expiration
// The signature and expiration are added as query parameters
func (g *SignedURLGenerator) GenerateSignedURL(baseURL string) (string, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse base URL: %w", err)
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(g.ttl).Unix()

	// Get existing query parameters
	queryParams := parsedURL.Query()
	queryParams.Set("expires", strconv.FormatInt(expiresAt, 10))

	// Create the string to sign (without signature)
	parsedURL.RawQuery = queryParams.Encode()
	stringToSign := parsedURL.String()

	// Generate signature
	signature := g.generateSignature(stringToSign)

	// Add signature to query parameters
	queryParams.Set("signature", signature)
	parsedURL.RawQuery = queryParams.Encode()

	return parsedURL.String(), nil
}

// VerifySignedURL verifies a signed URL's signature and expiration
func (g *SignedURLGenerator) VerifySignedURL(signedURL string) error {
	parsedURL, err := url.Parse(signedURL)
	if err != nil {
		return fmt.Errorf("failed to parse signed URL: %w", err)
	}

	queryParams := parsedURL.Query()

	// Check for required parameters
	expiresStr := queryParams.Get("expires")
	providedSignature := queryParams.Get("signature")

	if expiresStr == "" || providedSignature == "" {
		return ErrSignedURLMissingParams
	}

	// Check expiration
	expiresAt, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid expires parameter: %w", err)
	}

	if time.Now().Unix() > expiresAt {
		return ErrSignedURLExpired
	}

	// Verify signature
	// Remove signature from query params to get the original string to sign
	queryParams.Del("signature")
	parsedURL.RawQuery = queryParams.Encode()
	stringToSign := parsedURL.String()

	expectedSignature := g.generateSignature(stringToSign)

	if !hmac.Equal([]byte(providedSignature), []byte(expectedSignature)) {
		return ErrSignedURLInvalidSignature
	}

	return nil
}

// generateSignature creates an HMAC-SHA256 signature for the given data
func (g *SignedURLGenerator) generateSignature(data string) string {
	h := hmac.New(sha256.New, g.secretKey)
	h.Write([]byte(data))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return signature
}
