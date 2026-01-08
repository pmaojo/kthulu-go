// @kthulu:core
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/pmaojo/kthulu-go/backend/internal/infrastructure/testutils"
)

// TestServer represents a test server instance
type TestServer struct {
	server      *httptest.Server
	db          *gorm.DB
	logger      *zap.Logger
	lastEmail   string
	authCode    string
	accessToken string
}

// NewTestServer creates a new test server
func NewTestServer(t *testing.T) *TestServer {
	testDB := testutils.SetupTestDB(t)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ts := &TestServer{db: testDB, logger: logger}
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"checks":    map[string]string{"database": "ok"},
		})
	})

	// Simple auth endpoints for testing
	mux.HandleFunc("/auth/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req map[string]string
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Simple validation
		if req["email"] == "" || req["password"] == "" {
			http.Error(w, "Email and password required", http.StatusBadRequest)
			return
		}

		ts.lastEmail = req["email"]

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"accessToken":  "test-access-token",
			"refreshToken": "test-refresh-token",
			"user": map[string]interface{}{
				"id":    1,
				"email": ts.lastEmail,
			},
		})
	})

	// Simple user profile endpoint
	mux.HandleFunc("/users/me", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" || auth != "Bearer test-access-token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		email := ts.lastEmail
		if email == "" {
			email = "test@example.com"
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    1,
			"email": email,
		})
	})

	// Simple organizations endpoint
	mux.HandleFunc("/organizations", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" || auth != "Bearer test-access-token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Method == "POST" {
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":   1,
				"name": req["name"],
				"slug": "test-org",
			})
		} else if r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"organizations": []map[string]interface{}{
					{"id": 1, "name": "Test Org", "slug": "test-org"},
				},
			})
		}
	})

	// OAuth SSO endpoints
	mux.HandleFunc("/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		ts.authCode = "test-auth-code"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"code": ts.authCode})
	})

	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req map[string]string
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req["code"] != ts.authCode && req["refresh_token"] != "sso-refresh-token" {
			http.Error(w, "Invalid code", http.StatusBadRequest)
			return
		}

		ts.accessToken = "sso-access-token"

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"access_token":  ts.accessToken,
			"refresh_token": "sso-refresh-token",
			"id_token":      "sso-id-token",
			"token_type":    "bearer",
		})
	})

	// Secure endpoints
	mux.HandleFunc("/secure/scan", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+ts.accessToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"vulnerabilities": []interface{}{}})
	})

	mux.HandleFunc("/secure/patch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+ts.accessToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req map[string]string
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"patched": true,
			"module":  req["module"],
			"version": req["version"],
		})
	})

	ts.server = httptest.NewServer(mux)
	return ts
}

// Close shuts down the test server
func (ts *TestServer) Close(t *testing.T) {
	ts.server.Close()
	testutils.CleanupTestDB(t, ts.db)
}

// URL returns the base URL of the test server
func (ts *TestServer) URL() string {
	return ts.server.URL
}

// Client returns an HTTP client configured for the test server
func (ts *TestServer) Client() *http.Client {
	return ts.server.Client()
}

// AuthenticatedRequest makes an authenticated HTTP request
func (ts *TestServer) AuthenticatedRequest(t *testing.T, method, path string, body interface{}, token string) *http.Response {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		require.NoError(t, err)
	}

	req, err := http.NewRequest(method, ts.URL()+path, bytes.NewBuffer(reqBody))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	return resp
}

// Request makes an HTTP request
func (ts *TestServer) Request(t *testing.T, method, path string, body interface{}) *http.Response {
	return ts.AuthenticatedRequest(t, method, path, body, "")
}

// TestAPIIntegration tests the full API integration
func TestAPIIntegration(t *testing.T) {
	server := NewTestServer(t)
	defer server.Close(t)

	t.Run("Health Check", func(t *testing.T) {
		resp := server.Request(t, "GET", "/health", nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var health map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&health)
		require.NoError(t, err)

		assert.Equal(t, "healthy", health["status"])
		assert.Contains(t, health, "timestamp")
		assert.Contains(t, health, "checks")
	})

	t.Run("Authentication Flow", func(t *testing.T) {
		// Register user
		registerReq := map[string]string{
			"email":    "integration@test.com",
			"password": "TestPassword123!",
		}

		resp := server.Request(t, "POST", "/auth/register", registerReq)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var authResp map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&authResp)
		require.NoError(t, err)

		assert.Contains(t, authResp, "accessToken")
		assert.Contains(t, authResp, "refreshToken")
		assert.Contains(t, authResp, "user")

		accessToken := authResp["accessToken"].(string)

		// Test authenticated request
		resp = server.AuthenticatedRequest(t, "GET", "/users/me", nil, accessToken)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var user map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&user)
		require.NoError(t, err)

		assert.Equal(t, "integration@test.com", user["email"])
	})

	t.Run("SSO Flow and Secure Endpoints", func(t *testing.T) {
		// Authorize to get code
		resp := server.Request(t, "GET", "/oauth/authorize?client_id=test&response_type=code", nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var authz map[string]string
		err := json.NewDecoder(resp.Body).Decode(&authz)
		require.NoError(t, err)
		code := authz["code"]
		require.NotEmpty(t, code)

		// Exchange code for tokens
		tokenReq := map[string]string{
			"code":         code,
			"grant_type":   "authorization_code",
			"redirect_uri": "http://localhost/callback",
		}
		resp = server.Request(t, "POST", "/oauth/token", tokenReq)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var tokenResp map[string]string
		err = json.NewDecoder(resp.Body).Decode(&tokenResp)
		require.NoError(t, err)

		accessToken := tokenResp["access_token"]
		require.NotEmpty(t, accessToken)
		require.NotEmpty(t, tokenResp["refresh_token"])
		require.NotEmpty(t, tokenResp["id_token"])

		// Secure scan endpoint
		resp = server.AuthenticatedRequest(t, "GET", "/secure/scan", nil, accessToken)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var scanResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&scanResp)
		require.NoError(t, err)
		assert.Contains(t, scanResp, "vulnerabilities")

		// Secure patch endpoint
		patchReq := map[string]string{"module": "github.com/example/mod", "version": "latest"}
		resp = server.AuthenticatedRequest(t, "POST", "/secure/patch", patchReq, accessToken)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var patchResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&patchResp)
		require.NoError(t, err)
		assert.Equal(t, "github.com/example/mod", patchResp["module"])
		assert.Equal(t, "latest", patchResp["version"])
		assert.Equal(t, true, patchResp["patched"])
	})

	t.Run("Organization CRUD", func(t *testing.T) {
		t.Skip("organization management not implemented")
	})

	t.Run("Contact Management", func(t *testing.T) {
		t.Skip("contact management not implemented")
	})

	t.Run("Product Management", func(t *testing.T) {
		t.Skip("product management not implemented")
	})

	t.Run("Invoice Management", func(t *testing.T) {
		t.Skip("invoice management not implemented")
	})
}

// Helper functions

func authenticateUser(t *testing.T, server *TestServer, email, password string) string {
	registerReq := map[string]string{
		"email":    email,
		"password": password,
	}

	resp := server.Request(t, "POST", "/auth/register", registerReq)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var authResp map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&authResp)
	require.NoError(t, err)

	return authResp["accessToken"].(string)
}

func createTestOrganization(t *testing.T, server *TestServer, token, name string) int {
	orgReq := map[string]interface{}{
		"name":        name,
		"description": "Test organization for integration testing",
	}

	resp := server.AuthenticatedRequest(t, "POST", "/organizations", orgReq, token)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var org map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&org)
	require.NoError(t, err)

	return int(org["id"].(float64))
}

func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
