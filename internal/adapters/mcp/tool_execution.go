package mcp

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	mcp_golang "github.com/metoro-io/mcp-golang"
)

// NewReflectTool creates a RegisteredTool that uses reflection to map struct fields to CLI arguments.
// T must be a struct type with fields tagged with `kthulu:"..."`.
func NewReflectTool[T any](name, description string, baseArgs []string, executor CommandExecutor, workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        name,
		Description: description,
		Handler: func(ctx context.Context, args T) (*mcp_golang.ToolResponse, error) {
			resolved, err := resolveArgs(args, baseArgs)
			if err != nil {
				return nil, err
			}

			result, err := executor.Run(ctx, workingDir, resolved)
			commandLabel := strings.Join(append([]string{"kthulu"}, resolved...), " ")
			response := formatCommandResult(commandLabel, workingDir, result)

			if err != nil {
				return nil, fmt.Errorf("%s failed: %w\nSTDOUT:\n%s\nSTDERR:\n%s", commandLabel, err, normalizeOutput(result.Stdout), normalizeOutput(result.Stderr))
			}

			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(response)), nil
		},
	}
}

type posArg struct {
	index    int
	value    []string
	variadic bool
}

func resolveArgs(args any, baseArgs []string) ([]string, error) {
	v := reflect.ValueOf(args)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", v.Kind())
	}

	t := v.Type()

	var posArgs []posArg
	var flags []string

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		val := v.Field(i)
		tag := field.Tag.Get("kthulu")
		if tag == "" {
			continue
		}

		parts := strings.Split(tag, ",")
		kind := parts[0]

		meta := make(map[string]string)
		for _, p := range parts[1:] {
			kv := strings.SplitN(p, "=", 2)
			if len(kv) == 2 {
				meta[kv[0]] = kv[1]
			} else {
				meta[kv[0]] = "true"
			}
		}

		if kind == "pos" {
			idxStr := meta["index"]
			idx, _ := strconv.Atoi(idxStr)
			variadic := meta["variadic"] == "true"

			var values []string
			if variadic {
				if val.Kind() == reflect.Slice {
					for j := 0; j < val.Len(); j++ {
						values = append(values, val.Index(j).String())
					}
				}
			} else {
				s := val.String()
				if s != "" {
					values = append(values, s)
				}
			}

			if len(values) > 0 {
				posArgs = append(posArgs, posArg{index: idx, value: values, variadic: variadic})
			}
		} else if kind == "flag" {
			name := meta["name"]
			if name == "" {
				name = field.Name // fallback, though generator should provide
			}

			flagName := "--" + name

			switch val.Kind() {
			case reflect.Bool:
				if val.Bool() {
					flags = append(flags, flagName)
				}
			case reflect.Slice:
				for j := 0; j < val.Len(); j++ {
					flags = append(flags, flagName, val.Index(j).String())
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if val.Int() != 0 {
					flags = append(flags, flagName, fmt.Sprintf("%d", val.Int()))
				}
			case reflect.String:
				s := val.String()
				if s != "" {
					flags = append(flags, flagName, s)
				}
			}
		}
	}

	// Sort positional args
	sort.Slice(posArgs, func(i, j int) bool {
		return posArgs[i].index < posArgs[j].index
	})

	resolved := append([]string{}, baseArgs...)
	for _, p := range posArgs {
		resolved = append(resolved, p.value...)
	}
	resolved = append(resolved, flags...)

	return resolved, nil
}
