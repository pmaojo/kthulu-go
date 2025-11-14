package cmd

import (
	"io/fs"
	"os"
	"path/filepath"
)

// readlinkFS matches file systems that can read the target of a symbolic link.
// It is analogous to the fs.ReadlinkFS interface introduced in later Go
// versions and is used to replicate symlinks when available.
type readlinkFS interface {
	Readlink(name string) (string, error)
}

// copyFileFS copies a single file or symlink from fsys at src to dst on the
// local filesystem. File contents are read using fs.ReadFile and file modes are
// preserved. If fsys implements fs.ReadlinkFS and src is a symbolic link, the
// link is replicated. Existing files are skipped when skipExisting is true.
func copyFileFS(fsys fs.FS, src, dst string, skipExisting bool) error {
	if skipExisting {
		if _, err := os.Lstat(dst); err == nil {
			return nil
		}
	}
	if rlfs, ok := fsys.(readlinkFS); ok {
		if target, err := rlfs.Readlink(src); err == nil {
			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
				return err
			}
			return os.Symlink(target, dst)
		}
	}
	data, err := fs.ReadFile(fsys, src)
	if err != nil {
		return err
	}
	info, err := fs.Stat(fsys, src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, info.Mode())
}

// copyDirFS recursively copies a directory tree rooted at src from fsys to dst
// on the local filesystem. File contents are read using fs.ReadFile and file
// modes are preserved. If fsys implements fs.ReadlinkFS, symbolic links are
// replicated. Directories named .git or node_modules are skipped. Existing
// files are left untouched when skipExisting is true.
func copyDirFS(fsys fs.FS, src, dst string, skipExisting bool) error {
	return fs.WalkDir(fsys, src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := dst
		if rel != "." {
			target = filepath.Join(dst, rel)
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" {
				return fs.SkipDir
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			return os.MkdirAll(target, info.Mode())
		}
		return copyFileFS(fsys, path, target, skipExisting)
	})
}
