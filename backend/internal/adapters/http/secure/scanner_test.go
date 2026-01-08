package secure

import (
	"context"
	"testing"
)

func TestScanDeduplicates(t *testing.T) {
	// Mock dependencies
	origList := listModulesFn
	origQuery := queryOSVFn
	origRun := runGovulncheckFn
	defer func() {
		listModulesFn = origList
		queryOSVFn = origQuery
		runGovulncheckFn = origRun
	}()

	listModulesFn = func() ([]module, error) {
		return []module{{Path: "example/module", Version: "v1.0.0"}}, nil
	}
	queryOSVFn = func(ctx context.Context, path, version string) ([]Vuln, error) {
		return []Vuln{{Module: path, Version: version, ID: "CVE-1", Severity: "HIGH"}}, nil
	}
	runGovulncheckFn = func() ([]Vuln, error) {
		return []Vuln{{Module: "example/module", Version: "v1.0.0", ID: "CVE-1", Severity: "HIGH"}}, nil
	}

	vulns, err := Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if len(vulns) != 1 {
		t.Fatalf("expected 1 vulnerability, got %d", len(vulns))
	}
}
