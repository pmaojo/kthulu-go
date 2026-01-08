package builder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// Generate reads the plan at planPath and emits compiled.go and contracts_test.go in outDir.
// Generated files have build tag !nocli so they can be excluded with the nocli tag.
func Generate(planPath, outDir string) error {
	data, err := os.ReadFile(planPath)
	if err != nil {
		return err
	}
	var p Plan
	if len(data) > 0 {
		if err := json.Unmarshal(data, &p); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	compiled, err := buildCompiled(&p)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outDir, "compiled.go"), compiled, 0o644); err != nil {
		return err
	}
	if len(p.Replacements) > 0 {
		contracts, err := buildContracts(&p)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(outDir, "contracts_test.go"), contracts, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// buildCompiled generates the compiled.go file contents.
func buildCompiled(p *Plan) ([]byte, error) {
	// collect imports
	imports := map[string]string{"go.uber.org/fx": "fx"}
	aliasUsed := func(a string) bool {
		for _, v := range imports {
			if v == a {
				return true
			}
		}
		return false
	}
	aliasFor := func(pkg string) string {
		if a, ok := imports[pkg]; ok {
			return a
		}
		base := path.Base(pkg)
		alias := base
		i := 1
		for aliasUsed(alias) {
			alias = fmt.Sprintf("%s%d", base, i)
			i++
		}
		imports[pkg] = alias
		return alias
	}

	replaceCalls := []string{}
	for _, r := range p.Replacements {
		pkg, fn := splitPkgFn(r.Constructor)
		alias := aliasFor(pkg)
		replaceCalls = append(replaceCalls, fmt.Sprintf("%s.%s", alias, fn))
	}

	decorateCalls := []string{}
	for _, d := range p.Decorations {
		pkg, fn := splitPkgFn(d)
		alias := aliasFor(pkg)
		decorateCalls = append(decorateCalls, fmt.Sprintf("%s.%s", alias, fn))
	}

	groupLines := []string{}
	for grp, fns := range p.Groups {
		for _, fnPath := range fns {
			pkg, fn := splitPkgFn(fnPath)
			alias := aliasFor(pkg)
			groupLines = append(groupLines, fmt.Sprintf("fx.Annotate(%s.%s, fx.ResultTags(`group:\"%s\"`))", alias, fn, grp))
		}
	}

	// build file
	var buf bytes.Buffer
	buf.WriteString("//go:build !nocli\n")
	buf.WriteString("// +build !nocli\n\n")
	buf.WriteString("package overrides\n\n")
	buf.WriteString("import (\n")
	// sort imports for determinism
	pkgs := make([]string, 0, len(imports))
	for p := range imports {
		pkgs = append(pkgs, p)
	}
	sort.Strings(pkgs)
	for _, pkg := range pkgs {
		alias := imports[pkg]
		buf.WriteString(fmt.Sprintf("\t%s \"%s\"\n", alias, pkg))
	}
	buf.WriteString(")\n\n")
	buf.WriteString("var Module = fx.Options(\n")
	if len(replaceCalls) > 0 {
		buf.WriteString("\tfx.Replace(\n")
		for _, c := range replaceCalls {
			buf.WriteString("\t\t" + c + ",\n")
		}
		buf.WriteString("\t),\n")
	}
	if len(decorateCalls) > 0 {
		buf.WriteString("\tfx.Decorate(\n")
		for _, c := range decorateCalls {
			buf.WriteString("\t\t" + c + ",\n")
		}
		buf.WriteString("\t),\n")
	}
	if len(groupLines) > 0 {
		buf.WriteString("\tfx.Provide(\n")
		for _, line := range groupLines {
			buf.WriteString("\t\t" + line + ",\n")
		}
		buf.WriteString("\t),\n")
	}
	buf.WriteString(")\n")

	return format.Source(buf.Bytes())
}

// buildContracts generates contract tests ensuring implementations satisfy interfaces.
func buildContracts(p *Plan) ([]byte, error) {
	imports := map[string]string{}
	aliasUsed := func(a string) bool {
		for _, v := range imports {
			if v == a {
				return true
			}
		}
		return false
	}
	aliasFor := func(pkg string) string {
		if a, ok := imports[pkg]; ok {
			return a
		}
		base := path.Base(pkg)
		alias := base
		i := 1
		for aliasUsed(alias) {
			alias = fmt.Sprintf("%s%d", base, i)
			i++
		}
		imports[pkg] = alias
		return alias
	}

	lines := []string{}
	for _, r := range p.Replacements {
		ifacePkg, ifaceType := splitPkgFn(r.Interface)
		implPkg, implType := splitPkgFn(r.Implementation)
		ifaceAlias := aliasFor(ifacePkg)
		implAlias := aliasFor(implPkg)
		lines = append(lines, fmt.Sprintf("_ %s.%s = (*%s.%s)(nil)", ifaceAlias, ifaceType, implAlias, implType))
	}

	var buf bytes.Buffer
	buf.WriteString("//go:build !nocli\n")
	buf.WriteString("// +build !nocli\n\n")
	buf.WriteString("package overrides\n\n")
	buf.WriteString("import (\n")
	pkgs := make([]string, 0, len(imports))
	for p := range imports {
		pkgs = append(pkgs, p)
	}
	sort.Strings(pkgs)
	for _, pkg := range pkgs {
		alias := imports[pkg]
		buf.WriteString(fmt.Sprintf("\t%s \"%s\"\n", alias, pkg))
	}
	buf.WriteString(")\n\n")
	buf.WriteString("var (\n")
	for _, line := range lines {
		buf.WriteString("\t" + line + "\n")
	}
	buf.WriteString(")\n")

	return format.Source(buf.Bytes())
}

// splitPkgFn splits a qualified identifier like "pkg/path.Func" into the package path and name.
func splitPkgFn(q string) (string, string) {
	idx := strings.LastIndex(q, ".")
	if idx < 0 {
		return q, ""
	}
	return q[:idx], q[idx+1:]
}
