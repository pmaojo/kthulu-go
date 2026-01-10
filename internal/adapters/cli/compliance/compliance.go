package compliance

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CheckResult represents the outcome of a single compliance check
type CheckResult struct {
	Name        string
	Description string
	Passed      bool
	Error       error
}

// Report contains the full compliance validation report
type Report struct {
	Standard string
	Passed   bool
	Checks   []CheckResult
}

// Validator defines the interface for compliance checkers
type Validator interface {
	Validate(projectRoot string) (*Report, error)
}

// Validate performs compliance checks for the specified standard
func Validate(standard, projectRoot string) (*Report, error) {
	var v Validator

	switch strings.ToLower(standard) {
	case "sox":
		v = &SOXValidator{}
	case "gdpr":
		v = &GDPRValidator{}
	default:
		return nil, fmt.Errorf("unsupported compliance standard: %s", standard)
	}

	return v.Validate(projectRoot)
}

// SOXValidator implements Validator for Sarbanes-Oxley compliance
type SOXValidator struct{}

func (v *SOXValidator) Validate(root string) (*Report, error) {
	checks := []CheckResult{}

	// Check 1: Audit Logging
	// SOX requires detailed audit trails for financial data access
	auditCheck := CheckResult{
		Name:        "Audit Logging",
		Description: "Verify existence of audit logging mechanisms",
	}
	if found, _ := grepInFiles(root, "AuditLogger", "LogAudit", "audit_logs", "AuditMiddleware"); found {
		auditCheck.Passed = true
	} else {
		auditCheck.Error = fmt.Errorf("no audit logging mechanism found (searched for AuditLogger/LogAudit)")
	}
	checks = append(checks, auditCheck)

	// Check 2: Change Management (Migrations)
	// Changes to financial systems must be controlled and documented
	migrationCheck := CheckResult{
		Name:        "Change Management",
		Description: "Verify database migration system",
	}
	if exists(filepath.Join(root, "migrations")) || exists(filepath.Join(root, "db/migrations")) {
		migrationCheck.Passed = true
	} else {
		migrationCheck.Error = fmt.Errorf("migrations directory not found")
	}
	checks = append(checks, migrationCheck)

	// Check 3: Access Control
	// Verify RBAC implementation
	rbacCheck := CheckResult{
		Name:        "Access Control",
		Description: "Verify Role-Based Access Control (RBAC)",
	}
	if found, _ := grepInFiles(root, "RBAC", "Role", "Permission", "middleware.Auth", "Authorization"); found {
		rbacCheck.Passed = true
	} else {
		rbacCheck.Error = fmt.Errorf("no RBAC implementation traces found")
	}
	checks = append(checks, rbacCheck)

	passed := true
	for _, c := range checks {
		if !c.Passed {
			passed = false
			break
		}
	}

	return &Report{
		Standard: "SOX",
		Passed:   passed,
		Checks:   checks,
	}, nil
}

// GDPRValidator implements Validator for GDPR compliance
type GDPRValidator struct{}

func (v *GDPRValidator) Validate(root string) (*Report, error) {
	checks := []CheckResult{}

	// Check 1: Privacy Policy
	privacyCheck := CheckResult{
		Name:        "Privacy Policy",
		Description: "Check for privacy policy documentation",
	}
	// Check common filenames
	foundPrivacy := false
	for _, name := range []string{"PRIVACY.md", "privacy.md", "legal/privacy.md", "docs/privacy.md", "LEGAL.md"} {
		if exists(filepath.Join(root, name)) {
			foundPrivacy = true
			break
		}
	}
	if foundPrivacy {
		privacyCheck.Passed = true
	} else {
		privacyCheck.Error = fmt.Errorf("privacy policy file not found")
	}
	checks = append(checks, privacyCheck)

	// Check 2: Data Encryption
	cryptoCheck := CheckResult{
		Name:        "Data Encryption",
		Description: "Verify use of cryptography for sensitive data",
	}
	if found, _ := grepInFiles(root, "crypto/aes", "bcrypt", "argon2", "sslmode=require", "Encrypt", "Decrypt"); found {
		cryptoCheck.Passed = true
	} else {
		cryptoCheck.Error = fmt.Errorf("no standard encryption libraries or configurations found")
	}
	checks = append(checks, cryptoCheck)

	// Check 3: Right to be Forgotten
	rtbfCheck := CheckResult{
		Name:        "Right to be Forgotten",
		Description: "Verify data deletion capabilities",
	}
	if found, _ := grepInFiles(root, "DeleteUser", "Anonymize", "HardDelete", "PurgeData"); found {
		rtbfCheck.Passed = true
	} else {
		rtbfCheck.Error = fmt.Errorf("no user deletion/anonymization logic found")
	}
	checks = append(checks, rtbfCheck)

	passed := true
	for _, c := range checks {
		if !c.Passed {
			passed = false
			break
		}
	}

	return &Report{
		Standard: "GDPR",
		Passed:   passed,
		Checks:   checks,
	}, nil
}

// Helper functions

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

var ErrFound = errors.New("found")

func grepInFiles(root string, patterns ...string) (bool, error) {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			// skip hidden dirs like .git
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}
		// Only check source files and configs
		ext := filepath.Ext(path)
		if ext != ".go" && ext != ".yaml" && ext != ".json" && ext != ".sql" && ext != ".toml" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		strContent := string(content)
		for _, p := range patterns {
			if strings.Contains(strContent, p) {
				return ErrFound
			}
		}
		return nil
	})
	if err == ErrFound {
		return true, nil
	}
	return false, err
}
