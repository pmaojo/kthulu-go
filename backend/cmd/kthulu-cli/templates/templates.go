package templates

import "embed"

// Templates holds project scaffold files.
//
//go:embed backend frontend migrations openapi deployment.yaml deployment-info.txt scaffold
//go:embed go.mod.tmpl docker-compose.yml.tmpl Makefile.tmpl README.md.tmpl module.go.tmpl handler.go.tmpl handler_test.go.tmpl service_test.go.tmpl
var Templates embed.FS

// FS holds the new templates (alias to Templates for compatibility if everything is in root)
// Since I added frontend/admin/ into the same directory structure, Templates variable already includes it via `frontend` directive!
// "go:embed frontend" includes all subdirectories of frontend.
// So I don't need a new variable, I can just use `Templates`.
var FS = Templates
