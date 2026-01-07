// @kthulu:test:cmd:server
package main

import (
"context"
"net/http"
"net/http/httptest"
"testing"
"time"
)

type testHTTPServer struct {
started  bool
shutdown bool
}

func (s *testHTTPServer) Start() error {
s.started = true
return nil
}

func (s *testHTTPServer) Shutdown(context.Context) error {
s.shutdown = true
return nil
}

func TestRunApplicationLifecycle(t *testing.T) {
t.Setenv("KTHULU_TEST_MODE", "1")
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

srv := &testHTTPServer{}
errCh := make(chan error, 1)

go func() {
errCh <- runApplication(ctx, func(http.Handler) httpServer {
return srv
})
}()

time.Sleep(20 * time.Millisecond)
cancel()

select {
case err := <-errCh:
if err != nil {
t.Fatalf("runApplication returned error: %v", err)
}
case <-time.After(time.Second):
t.Fatal("timeout waiting for application shutdown")
}

if !srv.started || !srv.shutdown {
t.Fatalf("server lifecycle not executed: started=%v shutdown=%v", srv.started, srv.shutdown)
}
}

func TestSetupRoutesHealth(t *testing.T) {
handler := setupRoutes()
req := httptest.NewRequest(http.MethodGet, "/health", nil)
rr := httptest.NewRecorder()
handler.ServeHTTP(rr, req)

if rr.Code != http.StatusOK {
t.Fatalf("expected status 200 got %d", rr.Code)
}
if body := rr.Body.String(); body != "OK" {
t.Fatalf("expected OK body got %s", body)
}
}
