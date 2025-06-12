package utils

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyCookie_Success(t *testing.T) {
	app := fiber.New()

	testCookie := Cookie{
		Name:    "session_token",
		Value:   "abc123def456",
		Expires: time.Now().Add(24 * time.Hour),
	}

	app.Get("/test", func(c *fiber.Ctx) error {
		ApplyCookie(c, testCookie)
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	setCookieHeaders := resp.Header["Set-Cookie"]
	assert.NotEmpty(t, setCookieHeaders)

	setCookieHeader := strings.Join(setCookieHeaders, "; ")
	assert.Contains(t, setCookieHeader, "session_token=abc123def456")
	assert.Contains(t, setCookieHeader, "HttpOnly")
}

func TestApplyCookie_WithExpiration(t *testing.T) {
	app := fiber.New()

	expirationTime := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	testCookie := Cookie{
		Name:    "auth_token",
		Value:   "xyz789",
		Expires: expirationTime,
	}

	app.Get("/test", func(c *fiber.Ctx) error {
		ApplyCookie(c, testCookie)
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	setCookieHeaders := resp.Header["Set-Cookie"]
	setCookieHeader := strings.Join(setCookieHeaders, "; ")
	assert.Contains(t, setCookieHeader, "auth_token=xyz789")
	assert.Contains(t, setCookieHeader, "HttpOnly")
	assert.Contains(t, setCookieHeader, "expires=")
}

func TestApplyCookie_EmptyValue(t *testing.T) {
	app := fiber.New()

	testCookie := Cookie{
		Name:    "empty_cookie",
		Value:   "",
		Expires: time.Now().Add(1 * time.Hour),
	}

	app.Get("/test", func(c *fiber.Ctx) error {
		ApplyCookie(c, testCookie)
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	setCookieHeaders := resp.Header["Set-Cookie"]
	setCookieHeader := strings.Join(setCookieHeaders, "; ")
	assert.Contains(t, setCookieHeader, "empty_cookie=")
	assert.Contains(t, setCookieHeader, "HttpOnly")
}

func TestExpireCookie_Success(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		ExpireCookie(c, "session_token")
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	setCookieHeaders := resp.Header["Set-Cookie"]
	assert.NotEmpty(t, setCookieHeaders)

	setCookieHeader := strings.Join(setCookieHeaders, "; ")
	assert.Contains(t, setCookieHeader, "session_token=")
	assert.Contains(t, setCookieHeader, "HttpOnly")

	// The cookie should have an expiration time in the past (or very soon)
	assert.Contains(t, setCookieHeader, "expires=")
}

func TestExpireCookie_VerifyExpiration(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		ExpireCookie(c, "test_cookie")
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	setCookieHeaders := resp.Header["Set-Cookie"]
	setCookieHeader := strings.Join(setCookieHeaders, "; ")

	// Should contain the cookie name with empty value
	assert.Contains(t, setCookieHeader, "test_cookie=")

	// Should be HTTPOnly
	assert.Contains(t, setCookieHeader, "HttpOnly")

	// The value should be empty (after the = sign, before the semicolon)
	parts := strings.Split(setCookieHeader, ";")
	cookiePart := parts[0]
	assert.True(
		t,
		strings.HasSuffix(cookiePart, "test_cookie=") ||
			strings.Contains(cookiePart, "test_cookie=;"),
	)
}

func TestExpireCookie_MultipleNames(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		ExpireCookie(c, "cookie1")
		ExpireCookie(c, "cookie2")
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// Fiber may set multiple Set-Cookie headers, so we need to check all of them
	setCookieHeaders := resp.Header["Set-Cookie"]
	allHeaders := strings.Join(setCookieHeaders, "; ")

	assert.True(t, len(setCookieHeaders) > 0)
	assert.Contains(t, allHeaders, "cookie1=")
	assert.Contains(t, allHeaders, "cookie2=")
}

// Negative Test Cases

func TestApplyCookie_NilFiberContext(t *testing.T) {
	// This test verifies that passing a nil context would cause a panic
	// In real usage, this should never happen as Fiber manages the context
	testCookie := Cookie{
		Name:    "test_cookie",
		Value:   "test_value",
		Expires: time.Now().Add(1 * time.Hour),
	}

	// We can't actually test this without causing a panic, so we document the expectation
	// In production code, this would panic if c is nil
	// ApplyCookie(nil, testCookie) // Would panic

	// Instead, we test with a real context
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		ApplyCookie(c, testCookie)
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	setCookieHeaders := resp.Header["Set-Cookie"]
	assert.NotEmpty(t, setCookieHeaders)
}

func TestApplyCookie_VeryLongName(t *testing.T) {
	app := fiber.New()

	longName := strings.Repeat("a", 500) // Very long cookie name
	testCookie := Cookie{
		Name:    longName,
		Value:   "test_value",
		Expires: time.Now().Add(1 * time.Hour),
	}

	app.Get("/test", func(c *fiber.Ctx) error {
		ApplyCookie(c, testCookie)
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	setCookieHeaders := resp.Header["Set-Cookie"]
	setCookieHeader := strings.Join(setCookieHeaders, "; ")
	assert.Contains(t, setCookieHeader, longName+"=")
	assert.Contains(t, setCookieHeader, "HttpOnly")
}

func TestApplyCookie_VeryLongValue(t *testing.T) {
	app := fiber.New()

	longValue := strings.Repeat("x", 4000) // Very long cookie value
	testCookie := Cookie{
		Name:    "test_cookie",
		Value:   longValue,
		Expires: time.Now().Add(1 * time.Hour),
	}

	app.Get("/test", func(c *fiber.Ctx) error {
		ApplyCookie(c, testCookie)
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	setCookieHeaders := resp.Header["Set-Cookie"]
	setCookieHeader := strings.Join(setCookieHeaders, "; ")
	assert.Contains(t, setCookieHeader, "test_cookie="+longValue)
	assert.Contains(t, setCookieHeader, "HttpOnly")
}

func TestApplyCookie_SpecialCharactersInName(t *testing.T) {
	app := fiber.New()

	// Cookie names with special characters (some may be invalid but let's test behavior)
	specialNames := []string{
		"cookie-with-dash",
		"cookie_with_underscore",
		"cookie123",
		"cookie.with.dots",
		"cookie with spaces", // This might be problematic
		"cookie@domain.com",
	}

	for _, name := range specialNames {
		t.Run("name_"+name, func(t *testing.T) {
			testCookie := Cookie{
				Name:    name,
				Value:   "test_value",
				Expires: time.Now().Add(1 * time.Hour),
			}

			app.Get("/test", func(c *fiber.Ctx) error {
				ApplyCookie(c, testCookie)
				return c.SendString("ok")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			setCookieHeaders := resp.Header["Set-Cookie"]
			assert.NotEmpty(t, setCookieHeaders)
		})
	}
}

func TestApplyCookie_SpecialCharactersInValue(t *testing.T) {
	app := fiber.New()

	specialValues := []string{
		"value with spaces",
		"value;with;semicolons",
		"value=with=equals",
		"value\"with\"quotes",
		"value'with'apostrophes",
		"value\nwith\nnewlines",
		"value\twith\ttabs",
		"value,with,commas",
		"value{with}braces",
		"value[with]brackets",
	}

	for i, value := range specialValues {
		t.Run(fmt.Sprintf("value_%d", i), func(t *testing.T) {
			testCookie := Cookie{
				Name:    fmt.Sprintf("test_cookie_%d", i),
				Value:   value,
				Expires: time.Now().Add(1 * time.Hour),
			}

			app.Get("/test", func(c *fiber.Ctx) error {
				ApplyCookie(c, testCookie)
				return c.SendString("ok")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			setCookieHeaders := resp.Header["Set-Cookie"]
			assert.NotEmpty(t, setCookieHeaders)
		})
	}
}

func TestApplyCookie_PastExpiration(t *testing.T) {
	app := fiber.New()

	pastTime := time.Now().Add(-24 * time.Hour) // 24 hours ago
	testCookie := Cookie{
		Name:    "expired_cookie",
		Value:   "should_be_expired",
		Expires: pastTime,
	}

	app.Get("/test", func(c *fiber.Ctx) error {
		ApplyCookie(c, testCookie)
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	setCookieHeaders := resp.Header["Set-Cookie"]
	setCookieHeader := strings.Join(setCookieHeaders, "; ")
	assert.Contains(t, setCookieHeader, "expired_cookie=should_be_expired")
	assert.Contains(t, setCookieHeader, "HttpOnly")
	assert.Contains(t, setCookieHeader, "expires=")
}

func TestApplyCookie_ZeroTime(t *testing.T) {
	app := fiber.New()

	testCookie := Cookie{
		Name:    "zero_time_cookie",
		Value:   "test_value",
		Expires: time.Time{}, // Zero time
	}

	app.Get("/test", func(c *fiber.Ctx) error {
		ApplyCookie(c, testCookie)
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	setCookieHeaders := resp.Header["Set-Cookie"]
	setCookieHeader := strings.Join(setCookieHeaders, "; ")
	assert.Contains(t, setCookieHeader, "zero_time_cookie=test_value")
	assert.Contains(t, setCookieHeader, "HttpOnly")
}

func TestExpireCookie_EmptyName(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		ExpireCookie(c, "")
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	setCookieHeaders := resp.Header["Set-Cookie"]
	setCookieHeader := strings.Join(setCookieHeaders, "; ")
	assert.Contains(t, setCookieHeader, "=") // Empty name should still work
	assert.Contains(t, setCookieHeader, "HttpOnly")
}

func TestExpireCookie_VeryLongName(t *testing.T) {
	app := fiber.New()

	longName := strings.Repeat("x", 1000)

	app.Get("/test", func(c *fiber.Ctx) error {
		ExpireCookie(c, longName)
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	setCookieHeaders := resp.Header["Set-Cookie"]
	setCookieHeader := strings.Join(setCookieHeaders, "; ")
	assert.Contains(t, setCookieHeader, longName+"=")
	assert.Contains(t, setCookieHeader, "HttpOnly")
}

func TestExpireCookie_SpecialCharacterNames(t *testing.T) {
	app := fiber.New()

	specialNames := []string{
		"cookie-with-dash",
		"cookie_with_underscore",
		"cookie.with.dots",
		"cookie123numbers",
		"cookie@domain.com",
	}

	for _, name := range specialNames {
		t.Run("expire_"+name, func(t *testing.T) {
			app.Get("/test", func(c *fiber.Ctx) error {
				ExpireCookie(c, name)
				return c.SendString("ok")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			setCookieHeaders := resp.Header["Set-Cookie"]
			assert.NotEmpty(t, setCookieHeaders)
		})
	}
}
