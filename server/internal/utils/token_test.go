package utils

import (
	"fmt"
	"net/http/httptest"
	"server/config"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyToken(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		ApplyToken(c, "test-token-123")
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	authToken := resp.Header.Get("X-Auth-Token")
	assert.Equal(t, "test-token-123", authToken)
}

func TestGenerateJWTToken_Success(t *testing.T) {
	cfg := config.Config{
		SecurityJwtSecret: "test-secret-key-123",
	}

	userID := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)
	issuer := "test-app"

	token, err := GenerateJWTToken(userID, expiresAt, issuer, cfg)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Contains(t, token, ".")
}

func TestGenerateJWTToken_EmptySecret(t *testing.T) {
	cfg := config.Config{
		SecurityJwtSecret: "",
	}

	userID := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)
	issuer := "test-app"

	token, err := GenerateJWTToken(userID, expiresAt, issuer, cfg)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "JWT secret key not found")
}

func TestGenerateJWTToken_InvalidUserID(t *testing.T) {
	cfg := config.Config{
		SecurityJwtSecret: "test-secret-key-123",
	}

	invalidUserID := "not-a-uuid"
	expiresAt := time.Now().Add(24 * time.Hour)
	issuer := "test-app"

	token, err := GenerateJWTToken(invalidUserID, expiresAt, issuer, cfg)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "invalid UUID length")
}

func TestParseJWTToken_Success(t *testing.T) {
	cfg := config.Config{
		SecurityJwtSecret: "test-secret-key-123",
	}

	userID := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)
	issuer := "test-app"

	token, err := GenerateJWTToken(userID, expiresAt, issuer, cfg)
	require.NoError(t, err)

	claims, err := ParseJWTToken(token, cfg)

	require.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID.String())
	assert.Equal(t, issuer, claims.Issuer)
	assert.True(t, claims.ExpiresAt.After(time.Now()))
}

func TestParseJWTToken_EmptySecret(t *testing.T) {
	cfg := config.Config{
		SecurityJwtSecret: "",
	}

	token := "some.jwt.token"

	claims, err := ParseJWTToken(token, cfg)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "JWT secret key not found")
}

func TestParseJWTToken_InvalidToken(t *testing.T) {
	cfg := config.Config{
		SecurityJwtSecret: "test-secret-key-123",
	}

	invalidToken := "invalid.jwt.token"

	claims, err := ParseJWTToken(invalidToken, cfg)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestParseJWTToken_ExpiredToken(t *testing.T) {
	cfg := config.Config{
		SecurityJwtSecret: "test-secret-key-123",
	}

	userID := uuid.New().String()
	expiresAt := time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
	issuer := "test-app"

	token, err := GenerateJWTToken(userID, expiresAt, issuer, cfg)
	require.NoError(t, err)

	claims, err := ParseJWTToken(token, cfg)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestParseJWTToken_WrongSecret(t *testing.T) {
	cfg1 := config.Config{
		SecurityJwtSecret: "test-secret-key-123",
	}
	cfg2 := config.Config{
		SecurityJwtSecret: "different-secret-key",
	}

	userID := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)
	issuer := "test-app"

	token, err := GenerateJWTToken(userID, expiresAt, issuer, cfg1)
	require.NoError(t, err)

	claims, err := ParseJWTToken(token, cfg2)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

// Negative Test Cases

func TestApplyToken_EmptyToken(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		ApplyToken(c, "")
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	authToken := resp.Header.Get("X-Auth-Token")
	assert.Equal(t, "", authToken)
}

func TestApplyToken_SpecialCharacters(t *testing.T) {
	app := fiber.New()

	specialToken := "token-with-!@#$%^&*()_+={}[]|\\:;\"'<>?,./"
	app.Get("/test", func(c *fiber.Ctx) error {
		ApplyToken(c, specialToken)
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	authToken := resp.Header.Get("X-Auth-Token")
	assert.Equal(t, specialToken, authToken)
}

func TestGenerateJWTToken_EmptyUserID(t *testing.T) {
	cfg := config.Config{
		SecurityJwtSecret: "test-secret-key-123",
	}

	emptyUserID := ""
	expiresAt := time.Now().Add(24 * time.Hour)
	issuer := "test-app"

	token, err := GenerateJWTToken(emptyUserID, expiresAt, issuer, cfg)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "invalid UUID length")
}

func TestGenerateJWTToken_NilUUID(t *testing.T) {
	cfg := config.Config{
		SecurityJwtSecret: "test-secret-key-123",
	}

	nilUserID := "00000000-0000-0000-0000-000000000000"
	expiresAt := time.Now().Add(24 * time.Hour)
	issuer := "test-app"

	token, err := GenerateJWTToken(nilUserID, expiresAt, issuer, cfg)

	// This should succeed as nil UUID is still a valid UUID format
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify we can parse it back
	claims, err := ParseJWTToken(token, cfg)
	require.NoError(t, err)
	assert.Equal(t, nilUserID, claims.UserID.String())
}

func TestGenerateJWTToken_PastExpiration(t *testing.T) {
	cfg := config.Config{
		SecurityJwtSecret: "test-secret-key-123",
	}

	userID := uuid.New().String()
	expiresAt := time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
	issuer := "test-app"

	token, err := GenerateJWTToken(userID, expiresAt, issuer, cfg)

	// Generation should succeed even with past expiration
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// But parsing should fail due to expiration
	claims, err := ParseJWTToken(token, cfg)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "expired")
}

func TestGenerateJWTToken_EmptyIssuer(t *testing.T) {
	cfg := config.Config{
		SecurityJwtSecret: "test-secret-key-123",
	}

	userID := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)
	issuer := ""

	token, err := GenerateJWTToken(userID, expiresAt, issuer, cfg)

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify the empty issuer is preserved
	claims, err := ParseJWTToken(token, cfg)
	require.NoError(t, err)
	assert.Equal(t, "", claims.Issuer)
}

func TestParseJWTToken_EmptyToken(t *testing.T) {
	cfg := config.Config{
		SecurityJwtSecret: "test-secret-key-123",
	}

	claims, err := ParseJWTToken("", cfg)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestParseJWTToken_MalformedTokenStructure(t *testing.T) {
	cfg := config.Config{
		SecurityJwtSecret: "test-secret-key-123",
	}

	testCases := []struct {
		name  string
		token string
	}{
		{"no_dots", "notajwttoken"},
		{"one_dot", "header.payload"},
		{"too_many_dots", "header.payload.signature.extra"},
		{"empty_parts", ".."},
		{"spaces", "header . payload . signature"},
		{"invalid_base64", "!!!.!!!.!!!"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := ParseJWTToken(tc.token, cfg)
			assert.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

func TestParseJWTToken_ValidStructureInvalidSignature(t *testing.T) {
	cfg := config.Config{
		SecurityJwtSecret: "test-secret-key-123",
	}

	// Generate a valid token first
	userID := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)
	issuer := "test-app"

	validToken, err := GenerateJWTToken(userID, expiresAt, issuer, cfg)
	require.NoError(t, err)

	// Tamper with the signature part
	parts := strings.Split(validToken, ".")
	require.Len(t, parts, 3)

	tamperedToken := parts[0] + "." + parts[1] + ".tampered_signature"

	claims, err := ParseJWTToken(tamperedToken, cfg)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "signature is invalid")
}

func TestParseJWTToken_UnsupportedSigningMethod(t *testing.T) {
	cfg := config.Config{
		SecurityJwtSecret: "test-secret-key-123",
	}

	// Create a token with RS256 instead of HS256 (manual construction for testing)
	// Header: {"alg":"RS256","typ":"JWT"}
	// This should trigger the "unexpected signing method" error
	header := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9"
	payload := "eyJ1c2VySWQiOiIxMjM0NTY3OC05MDEyLTM0NTYtN2FiYy1kZWY0NTY3ODkwMTIiLCJpc3MiOiJ0ZXN0LWFwcCIsImV4cCI6MTk5OTk5OTk5OX0"
	signature := "invalid_signature"

	maliciousToken := header + "." + payload + "." + signature

	claims, err := ParseJWTToken(maliciousToken, cfg)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "unexpected signing method")
}

// Edge Case Tests

func TestGenerateJWTToken_ExtremelyShortSecret(t *testing.T) {
	cfg := config.Config{
		SecurityJwtSecret: "a", // Very short secret
	}

	userID := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)
	issuer := "test-app"

	token, err := GenerateJWTToken(userID, expiresAt, issuer, cfg)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Should still be parseable
	claims, err := ParseJWTToken(token, cfg)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID.String())
}

func TestGenerateJWTToken_VeryLongSecret(t *testing.T) {
	cfg := config.Config{
		SecurityJwtSecret: strings.Repeat("very-long-secret-key-", 100), // Very long secret
	}

	userID := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)
	issuer := "test-app"

	token, err := GenerateJWTToken(userID, expiresAt, issuer, cfg)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Should still be parseable
	claims, err := ParseJWTToken(token, cfg)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID.String())
}

func TestGenerateJWTToken_UnicodeInIssuer(t *testing.T) {
	cfg := config.Config{
		SecurityJwtSecret: "test-secret-key-123",
	}

	userID := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)
	issuer := "🚀 Test App 測試 ëxâmplé"

	token, err := GenerateJWTToken(userID, expiresAt, issuer, cfg)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ParseJWTToken(token, cfg)
	require.NoError(t, err)
	assert.Equal(t, issuer, claims.Issuer)
}

func TestParseJWTToken_ConcurrentAccess(t *testing.T) {
	cfg := config.Config{
		SecurityJwtSecret: "test-secret-key-123",
	}

	userID := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)
	issuer := "test-app"

	token, err := GenerateJWTToken(userID, expiresAt, issuer, cfg)
	require.NoError(t, err)

	// Test concurrent parsing of the same token
	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			claims, err := ParseJWTToken(token, cfg)
			if err != nil {
				results <- err
				return
			}
			if claims.UserID.String() != userID {
				results <- fmt.Errorf("userID mismatch: expected %s, got %s", userID, claims.UserID.String())
				return
			}
			results <- nil
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(t, err)
	}
}
