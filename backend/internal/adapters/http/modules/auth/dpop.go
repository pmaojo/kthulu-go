package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// ValidateDPoP verifies the DPoP proof for the request and binds it to the access token.
// It validates the HTTP method (htm), URL (htu) and access token hash (ath) claims.
// The proof's signature is not verified. This is a best-effort implementation.
func ValidateDPoP(r *http.Request, accessToken string) error {
	proof := r.Header.Get("DPoP")
	if proof == "" {
		return errors.New("missing DPoP header")
	}

	parts := strings.Split(proof, ".")
	if len(parts) < 2 {
		return errors.New("invalid DPoP format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return err
	}

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return err
	}

	htm, _ := claims["htm"].(string)
	htu, _ := claims["htu"].(string)
	ath, _ := claims["ath"].(string)

	if strings.ToUpper(htm) != strings.ToUpper(r.Method) {
		return errors.New("htm mismatch")
	}

	expectedHTU := r.URL.Scheme + "://" + r.Host + r.URL.Path
	if htu != expectedHTU {
		return errors.New("htu mismatch")
	}

	h := sha256.Sum256([]byte(accessToken))
	expectedATH := base64.RawURLEncoding.EncodeToString(h[:])
	if ath != expectedATH {
		return errors.New("ath mismatch")
	}

	return nil
}
