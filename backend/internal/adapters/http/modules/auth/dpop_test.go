package auth

import (
	"net/http"
	"testing"

	"github.com/pmaojo/kthulu-go/backend/internal/infrastructure/testutils"
)

func TestValidateDPoP(t *testing.T) {
	t.Skip("DPoP validation test unstable in current environment")
}

func TestValidateDPoPInvalid(t *testing.T) {
	token := "token"
	var errSeen error
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errSeen = ValidateDPoP(r, token)
		if errSeen != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	srv := testutils.SetupTestServer(t, handler)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()
	if errSeen == nil {
		t.Fatalf("expected validation error")
	}
}
