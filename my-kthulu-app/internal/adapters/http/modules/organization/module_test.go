// @kthulu:test:module:organization
package organization

import "testing"

func TestProviders(t *testing.T) {
if Providers() == nil {
t.Fatal("expected providers option")
}
}
