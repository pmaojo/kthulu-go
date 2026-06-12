package mcp

// pluginBuilders holds tool-builder functions registered via init() in any
// file in this package. This lets new tool files register themselves without
// touching factory.go, keeping parallel development conflict-free.
var pluginBuilders []func(executor CommandExecutor, workingDir string) []RegisteredTool

// RegisterPlugin adds a builder function to the plugin registry. Call this
// from init() in any tool file to make its tools appear in every MCP session.
func RegisterPlugin(b func(executor CommandExecutor, workingDir string) []RegisteredTool) {
	pluginBuilders = append(pluginBuilders, b)
}
