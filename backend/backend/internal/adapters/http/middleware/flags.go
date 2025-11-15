package middleware

import (
	"context"
	"net/http"

	flagcfg "github.com/kthulu/kthulu-go/backend/internal/modules/flags"
)

// FlagsKey is the context key for request flags.
const FlagsKey ContextKey = "flags"

// FlagsMiddleware extracts configured flags from headers and stores them in context.
func FlagsMiddleware(cfg flagcfg.HeaderConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(cfg) == 0 {
				next.ServeHTTP(w, r)
				return
			}
			flags := make(map[string]string)
			for header, name := range cfg {
				if val := r.Header.Get(header); val != "" {
					flags[name] = val
				}
			}
			ctx := context.WithValue(r.Context(), FlagsKey, flags)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetFlag returns a flag value from context.
func GetFlag(ctx context.Context, name string) (string, bool) {
	if flags, ok := ctx.Value(FlagsKey).(map[string]string); ok {
		val, ok := flags[name]
		return val, ok
	}
	return "", false
}

// GetAllFlags returns all flags stored in context.
func GetAllFlags(ctx context.Context) map[string]string {
	if flags, ok := ctx.Value(FlagsKey).(map[string]string); ok {
		return flags
	}
	return map[string]string{}
}
