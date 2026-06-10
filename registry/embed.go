// Package registry embeds the built-in marketplace registry (starters,
// modules and plugins) so 'kthulu marketplace' works on any machine without
// a checkout of the framework repository.
package registry

import "embed"

//go:embed starters modules plugins
var Files embed.FS
