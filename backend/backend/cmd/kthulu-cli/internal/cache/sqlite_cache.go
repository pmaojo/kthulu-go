package cache

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/internal/parser"
)

// Cache manages SQLite cache for tag analysis results
type Cache struct {
	db   *sql.DB
	path string
}

// NewCache creates a new cache instance
func NewCache(projectRoot string) (*Cache, error) {
	cachePath := filepath.Join(projectRoot, ".kthulu", "tags_cache.db")

	db, err := sql.Open("sqlite3", cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open cache database: %w", err)
	}

	cache := &Cache{
		db:   db,
		path: cachePath,
	}

	if err := cache.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	return cache, nil
}

// initialize creates the cache tables
func (c *Cache) initialize() error {
	schema := `
	CREATE TABLE IF NOT EXISTS file_hashes (
		file_path TEXT PRIMARY KEY,
		hash TEXT NOT NULL,
		last_modified INTEGER NOT NULL,
		created_at INTEGER NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS parsed_tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_path TEXT NOT NULL,
		tag_type TEXT NOT NULL,
		tag_value TEXT,
		tag_params TEXT, -- JSON
		package_name TEXT,
		context TEXT,
		line_number INTEGER,
		created_at INTEGER NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS modules (
		name TEXT PRIMARY KEY,
		package_name TEXT NOT NULL,
		dependencies TEXT, -- JSON array
		files TEXT, -- JSON array
		metadata TEXT, -- JSON
		last_updated INTEGER NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS dependencies (
		from_module TEXT NOT NULL,
		to_module TEXT NOT NULL,
		dependency_type TEXT NOT NULL,
		strength INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		PRIMARY KEY (from_module, to_module)
	);
	
	CREATE INDEX IF NOT EXISTS idx_tags_file_path ON parsed_tags(file_path);
	CREATE INDEX IF NOT EXISTS idx_tags_type ON parsed_tags(tag_type);
	CREATE INDEX IF NOT EXISTS idx_modules_package ON modules(package_name);
	`

	_, err := c.db.Exec(schema)
	return err
}

// IsFileCached checks if a file is already cached and up-to-date
func (c *Cache) IsFileCached(filePath string, fileHash string) (bool, error) {
	var cachedHash string
	var lastModified int64

	err := c.db.QueryRow(
		"SELECT hash, last_modified FROM file_hashes WHERE file_path = ?",
		filePath,
	).Scan(&cachedHash, &lastModified)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return cachedHash == fileHash, nil
}

// CacheFileTags caches parsed tags for a file
func (c *Cache) CacheFileTags(filePath, fileHash string, tags []parser.Tag) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().Unix()

	// Update file hash
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO file_hashes (file_path, hash, last_modified, created_at)
		VALUES (?, ?, ?, ?)
	`, filePath, fileHash, now, now)
	if err != nil {
		return err
	}

	// Delete old tags for this file
	_, err = tx.Exec("DELETE FROM parsed_tags WHERE file_path = ?", filePath)
	if err != nil {
		return err
	}

	// Insert new tags
	for _, tag := range tags {
		paramsJSON, _ := json.Marshal(tag.Params)

		_, err = tx.Exec(`
			INSERT INTO parsed_tags 
			(file_path, tag_type, tag_value, tag_params, package_name, context, line_number, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, filePath, string(tag.Type), tag.Value, string(paramsJSON), tag.Package, tag.Context, tag.Line, now)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetCachedTags retrieves cached tags for a file
func (c *Cache) GetCachedTags(filePath string) ([]parser.Tag, error) {
	rows, err := c.db.Query(`
		SELECT tag_type, tag_value, tag_params, package_name, context, line_number
		FROM parsed_tags WHERE file_path = ?
	`, filePath)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []parser.Tag

	for rows.Next() {
		var tag parser.Tag
		var paramsJSON string

		err := rows.Scan(
			&tag.Type, &tag.Value, &paramsJSON,
			&tag.Package, &tag.Context, &tag.Line,
		)
		if err != nil {
			return nil, err
		}

		tag.File = filePath
		tag.Params = make(map[string]string)
		if paramsJSON != "" {
			json.Unmarshal([]byte(paramsJSON), &tag.Params)
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

// CacheProjectAnalysis caches the complete project analysis
func (c *Cache) CacheProjectAnalysis(analysis *parser.ProjectAnalysis) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().Unix()

	// Clear old module data
	_, err = tx.Exec("DELETE FROM modules")
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM dependencies")
	if err != nil {
		return err
	}

	// Cache modules
	for _, module := range analysis.Modules {
		depsJSON, _ := json.Marshal(module.Dependencies)
		filesJSON, _ := json.Marshal(module.Files)
		metadataJSON, _ := json.Marshal(module.Metadata)

		_, err = tx.Exec(`
			INSERT INTO modules (name, package_name, dependencies, files, metadata, last_updated)
			VALUES (?, ?, ?, ?, ?, ?)
		`, module.Name, module.Package, string(depsJSON), string(filesJSON), string(metadataJSON), now)
		if err != nil {
			return err
		}
	}

	// Cache dependencies
	for _, dep := range analysis.Dependencies {
		_, err = tx.Exec(`
			INSERT INTO dependencies (from_module, to_module, dependency_type, strength, created_at)
			VALUES (?, ?, ?, ?, ?)
		`, dep.From, dep.To, dep.Type, dep.Strength, now)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetCachedModules retrieves all cached modules
func (c *Cache) GetCachedModules() (map[string]*parser.Module, error) {
	rows, err := c.db.Query(`
		SELECT name, package_name, dependencies, files, metadata
		FROM modules
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	modules := make(map[string]*parser.Module)

	for rows.Next() {
		var module parser.Module
		var depsJSON, filesJSON, metadataJSON string

		err := rows.Scan(
			&module.Name, &module.Package,
			&depsJSON, &filesJSON, &metadataJSON,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(depsJSON), &module.Dependencies)
		json.Unmarshal([]byte(filesJSON), &module.Files)
		json.Unmarshal([]byte(metadataJSON), &module.Metadata)

		modules[module.Name] = &module
	}

	return modules, nil
}

// GetCachedDependencies retrieves all cached dependencies
func (c *Cache) GetCachedDependencies() ([]parser.Dependency, error) {
	rows, err := c.db.Query(`
		SELECT from_module, to_module, dependency_type, strength
		FROM dependencies
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dependencies []parser.Dependency

	for rows.Next() {
		var dep parser.Dependency

		err := rows.Scan(&dep.From, &dep.To, &dep.Type, &dep.Strength)
		if err != nil {
			return nil, err
		}

		dependencies = append(dependencies, dep)
	}

	return dependencies, nil
}

// Close closes the cache database
func (c *Cache) Close() error {
	return c.db.Close()
}
