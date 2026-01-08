// @kthulu:core
package middleware

import (
	"net/http"

	"golang.org/x/time/rate"
)

// RateLimitMiddleware creates a middleware that limits the number of
// incoming requests using the provided rate limiter.
// If the limit is exceeded, the middleware responds with HTTP 429.
func RateLimitMiddleware(limiter *rate.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
