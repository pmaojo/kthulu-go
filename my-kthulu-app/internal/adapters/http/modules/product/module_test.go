// @kthulu:test:module:product
package product

import "testing"

func TestProviders(t *testing.T) {
if Providers() == nil {
t.Fatal("expected providers option")
}
}
