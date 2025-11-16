package secure

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os/exec"
	"strconv"
	"time"
)

// Vuln represents a discovered vulnerability.
type Vuln struct {
	Module   string
	Version  string
	ID       string
	Severity string
}

// allow dependency injection for tests
var (
	listModulesFn    = listModules
	queryOSVFn       = queryOSV
	runGovulncheckFn = runGovulncheck
)

// Scan inspects the current module and returns high severity vulnerabilities.
func Scan(ctx context.Context) ([]Vuln, error) {
	mods, err := listModulesFn()
	if err != nil {
		return nil, err
	}

	var vulns []Vuln
	for _, m := range mods {
		vs, err := queryOSVFn(ctx, m.Path, m.Version)
		if err != nil {
			return nil, err
		}
		vulns = append(vulns, vs...)
	}
	// attempt to run govulncheck for additional findings
	if gv, err := runGovulncheckFn(); err == nil {
		vulns = append(vulns, gv...)
	}

	uniq := map[string]Vuln{}
	for _, v := range vulns {
		key := v.Module + "@" + v.Version + ":" + v.ID
		uniq[key] = v
	}

	deduped := make([]Vuln, 0, len(uniq))
	for _, v := range uniq {
		deduped = append(deduped, v)
	}
	return deduped, nil
}

type module struct {
	Path    string
	Version string
}

func listModules() ([]module, error) {
	cmd := exec.Command("go", "list", "-m", "-json", "all")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(out))
	var mods []module
	for dec.More() {
		var m module
		if err := dec.Decode(&m); err != nil {
			return nil, err
		}
		mods = append(mods, m)
	}
	return mods, nil
}

func queryOSV(ctx context.Context, path, version string) ([]Vuln, error) {
	reqBody := map[string]any{
		"package": map[string]string{
			"name":      path,
			"ecosystem": "Go",
		},
		"version": version,
	}
	data, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "https://api.osv.dev/v1/query", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("osv query failed")
	}
	var body struct {
		Vulns []struct {
			ID       string   `json:"id"`
			Aliases  []string `json:"aliases"`
			Severity []struct {
				Score string `json:"score"`
			} `json:"severity"`
		} `json:"vulns"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	var vulns []Vuln
	for _, v := range body.Vulns {
		sev := classify(v.Severity)
		if sev == "LOW" || sev == "MEDIUM" || sev == "" {
			continue
		}
		id := v.ID
		for _, a := range v.Aliases {
			if len(a) > 4 && a[:4] == "CVE-" {
				id = a
				break
			}
		}
		vulns = append(vulns, Vuln{Module: path, Version: version, ID: id, Severity: sev})
	}
	return vulns, nil
}

func classify(sevs []struct {
	Score string `json:"score"`
}) string {
	var max float64
	for _, s := range sevs {
		if f, err := strconv.ParseFloat(s.Score, 64); err == nil {
			if f > max {
				max = f
			}
		}
	}
	switch {
	case max >= 9.0:
		return "CRITICAL"
	case max >= 7.0:
		return "HIGH"
	case max >= 4.0:
		return "MEDIUM"
	case max > 0:
		return "LOW"
	default:
		return ""
	}
}

func runGovulncheck() ([]Vuln, error) {
	if _, err := exec.LookPath("govulncheck"); err != nil {
		return nil, err
	}
	cmd := exec.Command("govulncheck", "-format=json", "./...")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	var vulns []Vuln
	for scanner.Scan() {
		line := scanner.Bytes()
		var m map[string]json.RawMessage
		if err := json.Unmarshal(line, &m); err != nil {
			continue
		}
		f, ok := m["finding"]
		if !ok {
			continue
		}
		var finding struct {
			OSV struct {
				ID       string   `json:"id"`
				Aliases  []string `json:"aliases"`
				Severity []struct {
					Score string `json:"score"`
				} `json:"severity"`
			} `json:"osv"`
			Module struct {
				Path    string `json:"path"`
				Version string `json:"version"`
			} `json:"module"`
		}
		if err := json.Unmarshal(f, &finding); err != nil {
			continue
		}
		sev := classify(finding.OSV.Severity)
		if sev == "LOW" || sev == "MEDIUM" || sev == "" {
			continue
		}
		id := finding.OSV.ID
		for _, a := range finding.OSV.Aliases {
			if len(a) > 4 && a[:4] == "CVE-" {
				id = a
				break
			}
		}
		vulns = append(vulns, Vuln{Module: finding.Module.Path, Version: finding.Module.Version, ID: id, Severity: sev})
	}
	return vulns, nil
}
