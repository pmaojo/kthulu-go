// @kthulu:test:module:users
package users

import "testing"

func TestProviders(t *testing.T) {
if Providers() == nil {
t.Fatal("expected providers option")
}
}
