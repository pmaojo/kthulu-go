package commands

import (
	"strings"
	"testing"

	"github.com/pmaojo/kthulu-go/registry"
)

func TestEmbeddedRegistryListsStarters(t *testing.T) {
	items, err := fetchMarketplaceItems(registry.Files)
	if err != nil {
		t.Fatal(err)
	}

	byID := map[string]MarketplaceItem{}
	for _, item := range items {
		byID[item.ID] = item
	}

	for _, id := range []string{"ecommerce-pro", "tournament-league", "crm-pipeline", "booking-engine", "blog-cms", "saas-starter"} {
		item, ok := byID[id]
		if !ok {
			t.Fatalf("starter %q missing from embedded registry", id)
		}
		if item.Type != "starter" {
			t.Fatalf("starter %q has type %q", id, item.Type)
		}
	}

	if byID["mcp-app-server"].Template != "mcp" {
		t.Fatalf("mcp-app-server must declare template=mcp, got %q", byID["mcp-app-server"].Template)
	}
}

func TestFindStarterPlanFromEmbeddedRegistry(t *testing.T) {
	plan, err := findStarterPlan(registry.Files, "ecommerce-pro")
	if err != nil {
		t.Fatal(err)
	}
	content := string(plan)
	for _, want := range []string{"modules:", "product:", "price:int:required,min=1", "buyer:belongs_to:customer"} {
		if !strings.Contains(content, want) {
			t.Fatalf("embedded plan missing %q:\n%s", want, content)
		}
	}

	// Legacy starters fall back to blueprint.yaml.
	plan, err = findStarterPlan(registry.Files, "saas-starter")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(plan), "subscription:") {
		t.Fatalf("saas blueprint missing modules:\n%s", plan)
	}

	if _, err := findStarterPlan(registry.Files, "does-not-exist"); err == nil {
		t.Fatal("expected error for unknown starter")
	}
}
