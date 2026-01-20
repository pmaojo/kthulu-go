package templates

import "embed"

// Templates holds project scaffold files.
//
//go:embed scaffold
var Templates embed.FS

// FS holds the templates for code generation.
var FS = Templates
