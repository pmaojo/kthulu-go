package mcp_test

import (
	"reflect"
	"testing"

	"github.com/invopop/jsonschema"
	"github.com/pmaojo/kthulu-go/backend/internal/adapters/mcp"
	"github.com/stretchr/testify/require"
)

func TestCommandArgumentsSchema(t *testing.T) {
	reflector := jsonschema.Reflector{
		Anonymous:                  true,
		AssignAnchor:               false,
		AllowAdditionalProperties:  true,
		RequiredFromJSONSchemaTags: true,
		DoNotReference:             true,
		ExpandedStruct:             true,
	}
	require.NotPanics(t, func() {
		reflector.ReflectFromType(reflect.TypeOf(mcp.CommandArguments{}))
	})
}
