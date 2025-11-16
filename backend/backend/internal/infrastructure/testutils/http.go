package testutils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// SetupTestServer starts an HTTP test server for the provided handler and
// registers a cleanup function. It simplifies spinning up temporary servers in
// unit tests.
func SetupTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}
