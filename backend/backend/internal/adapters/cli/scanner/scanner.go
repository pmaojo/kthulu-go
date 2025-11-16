package scanner

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Annotation represents a single override or extend directive discovered during scanning.
type Annotation struct {
	Module   string
	Symbol   string
	Priority int
	Mode     string
}

// Scan walks the overrides/ and extends/ directories rooted at base and returns
// all discovered annotations. An error is returned if duplicate annotations are
// encountered or if any annotation is invalid.
func Scan(base string) ([]Annotation, error) {
	dirs := []string{"overrides", "extends"}
	var result []Annotation
	seen := make(map[string]struct{})

	for _, d := range dirs {
		dir := filepath.Join(base, d)
		info, err := os.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		if !info.IsDir() {
			continue
		}

		err = filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			anns, err := parseFile(path)
			if err != nil {
				return err
			}
			for _, a := range anns {
				key := fmt.Sprintf("%s:%s:%s", a.Mode, a.Module, a.Symbol)
				if _, ok := seen[key]; ok {
					return fmt.Errorf("duplicate annotation %s", key)
				}
				seen[key] = struct{}{}
				result = append(result, a)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func parseFile(path string) ([]Annotation, error) {
	switch ext := strings.ToLower(filepath.Ext(path)); ext {
	case ".go":
		return parseGoFile(path)
	case ".yaml", ".yml":
		return parseYAMLFile(path)
	default:
		return nil, nil
	}
}

var tagRe = regexp.MustCompile(`@kthulu:(shadow|wrap)`)

func parseGoFile(path string) ([]Annotation, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	var anns []Annotation
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			if a, ok, err := parseAnnotationText(c.Text); ok {
				if err != nil {
					return nil, err
				}
				anns = append(anns, a)
			}
		}
	}
	return anns, nil
}

func parseAnnotationText(text string) (Annotation, bool, error) {
	text = strings.TrimSpace(strings.TrimPrefix(text, "//"))
	match := tagRe.FindStringSubmatch(text)
	if match == nil {
		return Annotation{}, false, nil
	}
	mode := match[1]
	rest := strings.TrimSpace(strings.Replace(text, match[0], "", 1))
	tokens := strings.Fields(rest)

	a := Annotation{Mode: mode}
	for _, tok := range tokens {
		kv := strings.SplitN(tok, ":", 2)
		if len(kv) != 2 {
			return Annotation{}, true, fmt.Errorf("invalid token %q", tok)
		}
		switch kv[0] {
		case "module":
			a.Module = kv[1]
		case "symbol":
			a.Symbol = kv[1]
		case "priority":
			p, err := strconv.Atoi(kv[1])
			if err != nil {
				return Annotation{}, true, fmt.Errorf("invalid priority %q", kv[1])
			}
			a.Priority = p
		default:
			return Annotation{}, true, fmt.Errorf("unknown field %q", kv[0])
		}
	}
	if a.Module == "" || a.Symbol == "" {
		return Annotation{}, true, fmt.Errorf("missing module or symbol")
	}
	return a, true, nil
}

type yamlAnnotation struct {
	Module   string `yaml:"module"`
	Symbol   string `yaml:"symbol"`
	Priority int    `yaml:"priority"`
}

func parseYAMLFile(path string) ([]Annotation, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var data struct {
		Kthulu map[string]yamlAnnotation `yaml:"kthulu"`
	}
	if err := yaml.Unmarshal(b, &data); err != nil {
		return nil, err
	}
	var anns []Annotation
	for mode, ann := range data.Kthulu {
		if mode != "shadow" && mode != "wrap" {
			return nil, fmt.Errorf("unknown mode %q", mode)
		}
		if ann.Module == "" || ann.Symbol == "" {
			return nil, fmt.Errorf("missing module or symbol for %s", mode)
		}
		anns = append(anns, Annotation{
			Module:   ann.Module,
			Symbol:   ann.Symbol,
			Priority: ann.Priority,
			Mode:     mode,
		})
	}
	return anns, nil
}
