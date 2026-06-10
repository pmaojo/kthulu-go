package mcp

import (
	"strings"
	"testing"
)

func TestReviewDomainModelFindsCommonMistakes(t *testing.T) {
	findings := reviewDomainModel([]ScaffoldModule{
		{Name: "teams", Fields: []string{"name:string", "created_date:string"}},
		{Name: "player", Fields: []string{"email:string", "team_id:int", "status:string"}},
	})

	joined := strings.Join(findings, "\n")
	for _, want := range []string{
		`use the singular form "team"`,
		"timestamps should use the time type",
		"add the email validation rule",
		"model this as a relation (team:belongs_to:team)",
		"constrain it with oneof",
		"no field is marked required",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing finding %q in:\n%s", want, joined)
		}
	}
}

func TestReviewDomainModelCleanModel(t *testing.T) {
	findings := reviewDomainModel([]ScaffoldModule{
		{Name: "team", Fields: []string{"name:string:required", "city:string"}},
		{Name: "player", Fields: []string{
			"name:string:required",
			"email:string:email",
			"joined_at:time",
			"squad:belongs_to:team",
		}},
	})
	if len(findings) != 0 {
		t.Fatalf("expected clean review, got: %v", findings)
	}
}

func TestReviewDomainModelFlagsMissingRelations(t *testing.T) {
	findings := reviewDomainModel([]ScaffoldModule{
		{Name: "team", Fields: []string{"name:string:required"}},
		{Name: "player", Fields: []string{"name:string:required"}},
	})
	joined := strings.Join(findings, "\n")
	if !strings.Contains(joined, "no relations between entities") {
		t.Fatalf("expected relation finding, got: %v", findings)
	}
}
