package cmd

import "testing"

func TestExportName(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"foo", "Foo"},
		{"foo bar", "Foo Bar"},
		{"foo_bar", "Foo_bar"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := exportName(tt.in); got != tt.want {
			t.Errorf("exportName(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
