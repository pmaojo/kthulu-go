// @kthulu:module:audit
package domain

import (
	"errors"
	"time"
)

// Domain errors
var (
	ErrInvalidAuditPath = errors.New("invalid audit path")
	ErrAuditNotFound    = errors.New("audit result not found")
)

// AuditRequest represents a request to perform code audit
type AuditRequest struct {
	Path       string   `json:"path,omitempty"`
	OnlyKinds  []string `json:"onlyKinds,omitempty"`
	Extensions []string `json:"extensions,omitempty"`
	Ignore     []string `json:"ignore,omitempty"`
	Strict     *bool    `json:"strict,omitempty"`
	Jobs       int      `json:"jobs,omitempty"`
}

// AuditResult represents the result of a code audit
type AuditResult struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Path      string         `json:"path"`
	Duration  string         `json:"duration"`
	Counts    map[string]int `json:"counts" gorm:"serializer:json"`
	Findings  []AuditFinding `json:"findings" gorm:"serializer:json"`
	Strict    bool           `json:"strict"`
	Warnings  []string       `json:"warnings,omitempty" gorm:"serializer:json"`
	CreatedAt time.Time      `json:"createdAt"`
}

// AuditFinding represents a single audit finding
type AuditFinding struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Kind   string `json:"kind"`
	Detail string `json:"detail"`
}

// NewAuditResult creates a new audit result
func NewAuditResult(path string, duration string, counts map[string]int, findings []AuditFinding, strict bool, warnings []string) *AuditResult {
	now := time.Now()
	return &AuditResult{
		Path:      path,
		Duration:  duration,
		Counts:    counts,
		Findings:  findings,
		Strict:    strict,
		Warnings:  warnings,
		CreatedAt: now,
	}
}

// AddFinding adds a finding to the audit result
func (a *AuditResult) AddFinding(finding AuditFinding) {
	a.Findings = append(a.Findings, finding)
}

// AddWarning adds a warning to the audit result
func (a *AuditResult) AddWarning(warning string) {
	a.Warnings = append(a.Warnings, warning)
}

// GetFindingsByKind returns findings filtered by kind
func (a *AuditResult) GetFindingsByKind(kind string) []AuditFinding {
	var filtered []AuditFinding
	for _, finding := range a.Findings {
		if finding.Kind == kind {
			filtered = append(filtered, finding)
		}
	}
	return filtered
}

// GetFindingsByFile returns findings filtered by file
func (a *AuditResult) GetFindingsByFile(file string) []AuditFinding {
	var filtered []AuditFinding
	for _, finding := range a.Findings {
		if finding.File == file {
			filtered = append(filtered, finding)
		}
	}
	return filtered
}

// HasErrors returns true if there are error-level findings
func (a *AuditResult) HasErrors() bool {
	for _, finding := range a.Findings {
		if finding.Kind == "error" {
			return true
		}
	}
	return false
}

// Summary returns a summary of the audit
func (a *AuditResult) Summary() map[string]int {
	summary := make(map[string]int)
	for _, finding := range a.Findings {
		summary[finding.Kind]++
	}
	return summary
}
