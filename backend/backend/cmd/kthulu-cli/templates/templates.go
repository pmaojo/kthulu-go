package templates

import "embed"

// Templates holds project scaffold files.
//
//go:embed backend frontend internal migrations openapi deployment.yaml deployment-info.txt
//go:embed go.mod.tmpl docker-compose.yml.tmpl Makefile.tmpl README.md.tmpl module.go.tmpl handler.go.tmpl handler_test.go.tmpl service_test.go.tmpl
var Templates embed.FS
