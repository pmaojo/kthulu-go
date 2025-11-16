package mcp_test

import (
	"testing"

	"github.com/pmaojo/kthulu-go/backend/internal/adapters/mcp"
	"github.com/stretchr/testify/require"
)

func TestAllowDenyFilterDenyTakesPrecedence(t *testing.T) {
	filter := mcp.NewAllowDenyFilter([]string{"deploy"}, []string{"deploy apply"})

	require.True(t, filter([]string{"deploy", "status"}))
	require.False(t, filter([]string{"deploy", "apply"}))
}

func TestAllowDenyFilterRestrictsWhenAllowListProvided(t *testing.T) {
	filter := mcp.NewAllowDenyFilter([]string{"status"}, nil)

	require.True(t, filter([]string{"status"}))
	require.False(t, filter([]string{"deploy"}))
}
