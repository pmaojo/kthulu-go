// @kthulu:test:module:auth
package auth

import "testing"

func TestProviders(t *testing.T) {
if Providers() == nil {
t.Fatal("expected providers option")
}
}
