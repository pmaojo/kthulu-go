package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	mcp_golang "github.com/metoro-io/mcp-golang"
)

// FileSystemService exposes direct file manipulation tools (read, write,
// surgical edits, listing, moving, and deleting) so agents can work on the
// project without shelling out.
type FileSystemService struct{}

// NewFileSystemService creates a new FileSystemService.
func NewFileSystemService() *FileSystemService {
	return &FileSystemService{}
}

// GetTools returns all filesystem tools bound to the working directory.
func (s *FileSystemService) GetTools(workingDir string) []RegisteredTool {
	return []RegisteredTool{
		s.readTool(workingDir),
		s.writeTool(workingDir),
		s.editTool(workingDir),
		s.listTool(workingDir),
		s.moveTool(workingDir),
		s.deleteTool(workingDir),
	}
}

// ReadFileArgs defines arguments for reading a file.
type ReadFileArgs struct {
	Path   string `json:"path" jsonschema:"description=File path relative to the project root (absolute paths allowed),required"`
	Offset int    `json:"offset,omitempty" jsonschema:"description=1-based line number to start reading from"`
	Limit  int    `json:"limit,omitempty" jsonschema:"description=Maximum number of lines to return (default: whole file)"`
}

// WriteFileArgs defines arguments for writing a file.
type WriteFileArgs struct {
	Path    string `json:"path" jsonschema:"description=File path relative to the project root,required"`
	Content string `json:"content" jsonschema:"description=Full content to write to the file,required"`
	Append  bool   `json:"append,omitempty" jsonschema:"description=Append to the file instead of overwriting"`
}

// EditFileArgs defines arguments for a surgical string replacement edit.
type EditFileArgs struct {
	Path       string `json:"path" jsonschema:"description=File path relative to the project root,required"`
	OldString  string `json:"old_string" jsonschema:"description=Exact text to replace (must match file content),required"`
	NewString  string `json:"new_string" jsonschema:"description=Replacement text"`
	ReplaceAll bool   `json:"replace_all,omitempty" jsonschema:"description=Replace every occurrence instead of requiring a unique match"`
}

// ListDirArgs defines arguments for listing a directory.
type ListDirArgs struct {
	Path      string `json:"path,omitempty" jsonschema:"description=Directory to list (default: project root)"`
	Recursive bool   `json:"recursive,omitempty" jsonschema:"description=Recurse into subdirectories"`
	MaxItems  int    `json:"max_items,omitempty" jsonschema:"description=Maximum entries to return (default 500)"`
}

// MoveFileArgs defines arguments for moving or renaming a file.
type MoveFileArgs struct {
	Source      string `json:"source" jsonschema:"description=Current path of the file or directory,required"`
	Destination string `json:"destination" jsonschema:"description=New path for the file or directory,required"`
}

// DeleteFileArgs defines arguments for deleting a file or directory.
type DeleteFileArgs struct {
	Path      string `json:"path" jsonschema:"description=Path of the file or directory to delete,required"`
	Recursive bool   `json:"recursive,omitempty" jsonschema:"description=Required to delete non-empty directories"`
}

func (s *FileSystemService) readTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "fs_read",
		Description: "Read the content of a file. Supports reading a slice of the file via offset/limit (line based).",
		Handler: func(ctx context.Context, args ReadFileArgs) (*mcp_golang.ToolResponse, error) {
			content, err := s.ReadFile(workingDir, args)
			if err != nil {
				return nil, err
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(content)), nil
		},
	}
}

func (s *FileSystemService) writeTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "fs_write",
		Description: "Create or overwrite a file with the given content. Parent directories are created automatically.",
		Handler: func(ctx context.Context, args WriteFileArgs) (*mcp_golang.ToolResponse, error) {
			summary, err := s.WriteFile(workingDir, args)
			if err != nil {
				return nil, err
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(summary)), nil
		},
	}
}

func (s *FileSystemService) editTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "fs_edit",
		Description: "Perform an exact string replacement in a file. old_string must match the file content exactly; if it matches more than once, set replace_all or provide more surrounding context.",
		Handler: func(ctx context.Context, args EditFileArgs) (*mcp_golang.ToolResponse, error) {
			summary, err := s.EditFile(workingDir, args)
			if err != nil {
				return nil, err
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(summary)), nil
		},
	}
}

func (s *FileSystemService) listTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "fs_list",
		Description: "List files and directories. Directories are suffixed with '/'. Common build/VCS directories are skipped when recursing.",
		Handler: func(ctx context.Context, args ListDirArgs) (*mcp_golang.ToolResponse, error) {
			listing, err := s.ListDir(workingDir, args)
			if err != nil {
				return nil, err
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(listing)), nil
		},
	}
}

func (s *FileSystemService) moveTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "fs_move",
		Description: "Move or rename a file or directory. Parent directories of the destination are created automatically.",
		Handler: func(ctx context.Context, args MoveFileArgs) (*mcp_golang.ToolResponse, error) {
			summary, err := s.MoveFile(workingDir, args)
			if err != nil {
				return nil, err
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(summary)), nil
		},
	}
}

func (s *FileSystemService) deleteTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "fs_delete",
		Description: "Delete a file or directory. Deleting a non-empty directory requires recursive=true.",
		Handler: func(ctx context.Context, args DeleteFileArgs) (*mcp_golang.ToolResponse, error) {
			summary, err := s.DeleteFile(workingDir, args)
			if err != nil {
				return nil, err
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(summary)), nil
		},
	}
}

// ReadFile reads a file, optionally slicing it by line offset and limit.
func (s *FileSystemService) ReadFile(workingDir string, args ReadFileArgs) (string, error) {
	path, err := resolveWorkspacePath(workingDir, args.Path)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", args.Path, err)
	}

	content := string(data)
	if args.Offset <= 0 && args.Limit <= 0 {
		return truncateOutput(content), nil
	}

	lines := strings.Split(content, "\n")
	start := args.Offset
	if start <= 0 {
		start = 1
	}
	if start > len(lines) {
		return "", fmt.Errorf("offset %d is beyond the end of the file (%d lines)", start, len(lines))
	}

	end := len(lines)
	if args.Limit > 0 && start-1+args.Limit < end {
		end = start - 1 + args.Limit
	}

	slice := strings.Join(lines[start-1:end], "\n")
	header := fmt.Sprintf("%s (lines %d-%d of %d)\n", args.Path, start, end, len(lines))
	return truncateOutput(header + slice), nil
}

// WriteFile creates or overwrites a file, creating parent directories as needed.
func (s *FileSystemService) WriteFile(workingDir string, args WriteFileArgs) (string, error) {
	path, err := resolveWorkspacePath(workingDir, args.Path)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("failed to create parent directories for %s: %w", args.Path, err)
	}

	if args.Append {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return "", fmt.Errorf("failed to open %s for append: %w", args.Path, err)
		}
		defer f.Close()
		if _, err := f.WriteString(args.Content); err != nil {
			return "", fmt.Errorf("failed to append to %s: %w", args.Path, err)
		}
		return fmt.Sprintf("Appended %d bytes to %s", len(args.Content), args.Path), nil
	}

	if err := os.WriteFile(path, []byte(args.Content), 0o644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", args.Path, err)
	}
	return fmt.Sprintf("Wrote %d bytes to %s", len(args.Content), args.Path), nil
}

// EditFile performs an exact string replacement inside a file.
func (s *FileSystemService) EditFile(workingDir string, args EditFileArgs) (string, error) {
	if args.OldString == "" {
		return "", fmt.Errorf("argument 'old_string' is required")
	}
	if args.OldString == args.NewString {
		return "", fmt.Errorf("old_string and new_string are identical")
	}

	path, err := resolveWorkspacePath(workingDir, args.Path)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", args.Path, err)
	}
	content := string(data)

	count := strings.Count(content, args.OldString)
	if count == 0 {
		return "", fmt.Errorf("old_string not found in %s", args.Path)
	}
	if count > 1 && !args.ReplaceAll {
		return "", fmt.Errorf("old_string matches %d times in %s; provide more surrounding context or set replace_all=true", count, args.Path)
	}

	var updated string
	replaced := count
	if args.ReplaceAll {
		updated = strings.ReplaceAll(content, args.OldString, args.NewString)
	} else {
		updated = strings.Replace(content, args.OldString, args.NewString, 1)
		replaced = 1
	}

	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", args.Path, err)
	}

	line := 1 + strings.Count(content[:strings.Index(content, args.OldString)], "\n")
	return fmt.Sprintf("Replaced %d occurrence(s) in %s (first match at line %d)", replaced, args.Path, line), nil
}

// ListDir lists directory entries, optionally recursively.
func (s *FileSystemService) ListDir(workingDir string, args ListDirArgs) (string, error) {
	relPath := args.Path
	if strings.TrimSpace(relPath) == "" {
		relPath = "."
	}
	path, err := resolveWorkspacePath(workingDir, relPath)
	if err != nil {
		return "", err
	}

	maxItems := args.MaxItems
	if maxItems <= 0 {
		maxItems = 500
	}

	var entries []string
	if args.Recursive {
		err = filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if p == path {
				return nil
			}
			if d.IsDir() && shouldSkipDir(d.Name()) {
				return filepath.SkipDir
			}
			rel, relErr := filepath.Rel(path, p)
			if relErr != nil {
				rel = p
			}
			if d.IsDir() {
				rel += "/"
			}
			entries = append(entries, rel)
			if len(entries) >= maxItems {
				return filepath.SkipAll
			}
			return nil
		})
		if err != nil {
			return "", fmt.Errorf("failed to walk %s: %w", relPath, err)
		}
	} else {
		dirEntries, err := os.ReadDir(path)
		if err != nil {
			return "", fmt.Errorf("failed to list %s: %w", relPath, err)
		}
		for _, entry := range dirEntries {
			name := entry.Name()
			if entry.IsDir() {
				name += "/"
			}
			entries = append(entries, name)
			if len(entries) >= maxItems {
				break
			}
		}
	}

	if len(entries) == 0 {
		return fmt.Sprintf("%s is empty", relPath), nil
	}

	sort.Strings(entries)
	header := fmt.Sprintf("%s (%d entries)\n", relPath, len(entries))
	return truncateOutput(header + strings.Join(entries, "\n")), nil
}

// MoveFile moves or renames a file or directory.
func (s *FileSystemService) MoveFile(workingDir string, args MoveFileArgs) (string, error) {
	source, err := resolveWorkspacePath(workingDir, args.Source)
	if err != nil {
		return "", fmt.Errorf("invalid source: %w", err)
	}
	destination, err := resolveWorkspacePath(workingDir, args.Destination)
	if err != nil {
		return "", fmt.Errorf("invalid destination: %w", err)
	}

	if _, err := os.Stat(source); err != nil {
		return "", fmt.Errorf("source %s is not accessible: %w", args.Source, err)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return "", fmt.Errorf("failed to create parent directories for %s: %w", args.Destination, err)
	}
	if err := os.Rename(source, destination); err != nil {
		return "", fmt.Errorf("failed to move %s to %s: %w", args.Source, args.Destination, err)
	}
	return fmt.Sprintf("Moved %s to %s", args.Source, args.Destination), nil
}

// DeleteFile removes a file or directory.
func (s *FileSystemService) DeleteFile(workingDir string, args DeleteFileArgs) (string, error) {
	path, err := resolveWorkspacePath(workingDir, args.Path)
	if err != nil {
		return "", err
	}

	cleanWorkingDir := filepath.Clean(workingDir)
	if path == cleanWorkingDir || path == string(filepath.Separator) {
		return "", fmt.Errorf("refusing to delete the project root")
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("%s is not accessible: %w", args.Path, err)
	}

	if info.IsDir() {
		if !args.Recursive {
			if err := os.Remove(path); err != nil {
				return "", fmt.Errorf("failed to delete directory %s (set recursive=true for non-empty directories): %w", args.Path, err)
			}
		} else if err := os.RemoveAll(path); err != nil {
			return "", fmt.Errorf("failed to delete directory %s: %w", args.Path, err)
		}
		return fmt.Sprintf("Deleted directory %s", args.Path), nil
	}

	if err := os.Remove(path); err != nil {
		return "", fmt.Errorf("failed to delete %s: %w", args.Path, err)
	}
	return fmt.Sprintf("Deleted %s", args.Path), nil
}

// resolveWorkspacePath resolves a tool-provided path relative to the working
// directory. Absolute paths are accepted as-is, matching the power level of
// the shell skill.
func resolveWorkspacePath(workingDir, path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("argument 'path' is required")
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	return filepath.Clean(filepath.Join(workingDir, path)), nil
}

// shouldSkipDir reports whether a directory is noise for listing/search/AST walks.
func shouldSkipDir(name string) bool {
	switch name {
	case ".git", ".idea", ".vscode", "node_modules", "vendor", "dist", "bin", "tmp", ".cache":
		return true
	}
	return false
}

const maxToolOutputBytes = 50000

// truncateOutput caps tool output at a safe size for MCP responses.
func truncateOutput(output string) string {
	if len(output) > maxToolOutputBytes {
		return output[:maxToolOutputBytes] + "\n...[Output Truncated]..."
	}
	return output
}
