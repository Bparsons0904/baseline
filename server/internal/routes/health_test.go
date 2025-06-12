package routes

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"server/config"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthRoutes(t *testing.T) {
	testConfig := config.Config{
		GeneralVersion: "1.2.3",
	}

	app := fiber.New()
	HealthRoutes(app, testConfig)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var healthResponse map[string]interface{}
	err = json.Unmarshal(body, &healthResponse)
	require.NoError(t, err)

	assert.Equal(t, "ok", healthResponse["status"])
	assert.Equal(t, "1.2.3", healthResponse["version"])
	assert.Equal(t, "app_api", healthResponse["service"])
}

func TestHealthRoutes_WithEmptyVersion(t *testing.T) {
	testConfig := config.Config{
		GeneralVersion: "",
	}

	app := fiber.New()
	HealthRoutes(app, testConfig)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var healthResponse map[string]interface{}
	err = json.Unmarshal(body, &healthResponse)
	require.NoError(t, err)

	assert.Equal(t, "ok", healthResponse["status"])
	assert.Equal(t, "", healthResponse["version"])
	assert.Equal(t, "app_api", healthResponse["service"])
}

func TestHealthRoutes_ResponseStructure(t *testing.T) {
	testConfig := config.Config{
		GeneralVersion: "test-version",
	}

	app := fiber.New()
	HealthRoutes(app, testConfig)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var healthResponse map[string]interface{}
	err = json.Unmarshal(body, &healthResponse)
	require.NoError(t, err)

	// Verify all required fields are present
	assert.Contains(t, healthResponse, "status")
	assert.Contains(t, healthResponse, "version")
	assert.Contains(t, healthResponse, "service")

	// Verify field types
	assert.IsType(t, "", healthResponse["status"])
	assert.IsType(t, "", healthResponse["version"])
	assert.IsType(t, "", healthResponse["service"])
}

func TestHealthRoutes_ContentType(t *testing.T) {
	testConfig := config.Config{
		GeneralVersion: "1.0.0",
	}

	app := fiber.New()
	HealthRoutes(app, testConfig)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	contentType := resp.Header.Get("Content-Type")
	assert.Contains(t, contentType, "application/json")
}

func TestHealthRoutes_HTTPMethods(t *testing.T) {
	testConfig := config.Config{
		GeneralVersion: "1.0.0",
	}

	app := fiber.New()
	HealthRoutes(app, testConfig)

	// Test GET method works
	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Test POST method should return method not allowed
	req = httptest.NewRequest("POST", "/health", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusMethodNotAllowed, resp.StatusCode)

	// Test PUT method should return method not allowed
	req = httptest.NewRequest("PUT", "/health", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusMethodNotAllowed, resp.StatusCode)

	// Test DELETE method should return method not allowed
	req = httptest.NewRequest("DELETE", "/health", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusMethodNotAllowed, resp.StatusCode)
}

func TestHealthRoutes_MultipleRequests(t *testing.T) {
	testConfig := config.Config{
		GeneralVersion: "1.0.0",
	}

	app := fiber.New()
	HealthRoutes(app, testConfig)

	// Make multiple requests to ensure consistency
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var healthResponse map[string]interface{}
		err = json.Unmarshal(body, &healthResponse)
		require.NoError(t, err)

		assert.Equal(t, "ok", healthResponse["status"])
		assert.Equal(t, "1.0.0", healthResponse["version"])
		assert.Equal(t, "app_api", healthResponse["service"])
	}
}

func TestHealthRoutes_ConfigVariations(t *testing.T) {
	testCases := []struct {
		name            string
		version         string
		expectedVersion string
	}{
		{
			name:            "normal version",
			version:         "1.2.3",
			expectedVersion: "1.2.3",
		},
		{
			name:            "semantic version",
			version:         "1.2.3-beta.1",
			expectedVersion: "1.2.3-beta.1",
		},
		{
			name:            "simple version",
			version:         "v1",
			expectedVersion: "v1",
		},
		{
			name:            "empty version",
			version:         "",
			expectedVersion: "",
		},
		{
			name:            "special characters",
			version:         "1.2.3+build.123",
			expectedVersion: "1.2.3+build.123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testConfig := config.Config{
				GeneralVersion: tc.version,
			}

			app := fiber.New()
			HealthRoutes(app, testConfig)

			req := httptest.NewRequest("GET", "/health", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var healthResponse map[string]interface{}
			err = json.Unmarshal(body, &healthResponse)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedVersion, healthResponse["version"])
		})
	}
}
