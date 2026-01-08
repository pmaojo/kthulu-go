package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// AdvancedTagParser provides enhanced parsing capabilities
type AdvancedTagParser struct {
	cache    Cache
	fileSet  *token.FileSet
	insights *SemanticInsights
	depGraph *DependencyGraph
	mutex    sync.RWMutex
	config   *AdvancedParserConfig
}

// AdvancedParserConfig configures the advanced parser
type AdvancedParserConfig struct {
	CacheEnabled      bool          `json:"cache_enabled"`
	CacheTTL          time.Duration `json:"cache_ttl"`
	CircularDetection bool          `json:"circular_detection"`
	SemanticAnalysis  bool          `json:"semantic_analysis"`
	MaxFileSize       int64         `json:"max_file_size"`
	IgnorePatterns    []string      `json:"ignore_patterns"`
}

// SemanticInsights provides semantic analysis results
type SemanticInsights struct {
	Patterns        map[string]*CodePattern `json:"patterns"`
	Recommendations []Recommendation        `json:"recommendations"`
	Metrics         *ProjectMetrics         `json:"metrics"`
	LastAnalyzed    time.Time               `json:"last_analyzed"`
}

// CodePattern represents a detected code pattern
type CodePattern struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Occurrences int               `json:"occurrences"`
	Files       []string          `json:"files"`
	Confidence  float64           `json:"confidence"`
	Metadata    map[string]string `json:"metadata"`
}

// Recommendation provides actionable insights
type Recommendation struct {
	Type        string            `json:"type"`
	Severity    string            `json:"severity"`
	Message     string            `json:"message"`
	File        string            `json:"file"`
	Line        int               `json:"line"`
	Suggestions []string          `json:"suggestions"`
	Metadata    map[string]string `json:"metadata"`
}

// ProjectMetrics contains project-level metrics
type ProjectMetrics struct {
	TotalFiles      int                       `json:"total_files"`
	TotalLines      int                       `json:"total_lines"`
	TaggedFiles     int                       `json:"tagged_files"`
	ModuleCount     int                       `json:"module_count"`
	CycleCount      int                       `json:"cycle_count"`
	ComplexityScore float64                   `json:"complexity_score"`
	TagDistribution map[TagType]int           `json:"tag_distribution"`
	DependencyDepth int                       `json:"dependency_depth"`
	ModuleMetrics   map[string]*ModuleMetrics `json:"module_metrics"`
}

// ModuleMetrics contains module-level metrics
type ModuleMetrics struct {
	FileCount       int     `json:"file_count"`
	LineCount       int     `json:"line_count"`
	TagCount        int     `json:"tag_count"`
	DependencyCount int     `json:"dependency_count"`
	CoverageScore   float64 `json:"coverage_score"`
	QualityScore    float64 `json:"quality_score"`
}

// DependencyGraph represents the project dependency graph
type DependencyGraph struct {
	Nodes  map[string]*DependencyNode `json:"nodes"`
	Edges  []*DependencyEdge          `json:"edges"`
	Cycles [][]string                 `json:"cycles"`
	Levels map[string]int             `json:"levels"`
}

// DependencyNode represents a node in the dependency graph
type DependencyNode struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Tags      []Tag             `json:"tags"`
	Metadata  map[string]string `json:"metadata"`
	InDegree  int               `json:"in_degree"`
	OutDegree int               `json:"out_degree"`
}

// DependencyEdge represents an edge in the dependency graph
type DependencyEdge struct {
	From     string            `json:"from"`
	To       string            `json:"to"`
	Type     string            `json:"type"`
	Weight   int               `json:"weight"`
	Metadata map[string]string `json:"metadata"`
}

// MemoryCache implements a simple in-memory cache
type MemoryCache struct {
	data   map[string]*CacheEntry
	mutex  sync.RWMutex
	maxTTL time.Duration
}

// CacheEntry represents a cached entry
type CacheEntry struct {
	Data      []byte        `json:"data"`
	CreatedAt time.Time     `json:"created_at"`
	TTL       time.Duration `json:"ttl"`
}

// NewAdvancedTagParser creates a new advanced tag parser
func NewAdvancedTagParser(config *AdvancedParserConfig) *AdvancedTagParser {
	if config == nil {
		config = &AdvancedParserConfig{
			CacheEnabled:      true,
			CacheTTL:          time.Hour,
			CircularDetection: true,
			SemanticAnalysis:  true,
			MaxFileSize:       10 * 1024 * 1024, // 10MB
		}
	}

	cache := NewMemoryCache(config.CacheTTL)
	if !config.CacheEnabled {
		cache = &MemoryCache{data: make(map[string]*CacheEntry), maxTTL: time.Minute} // Minimal cache
	}

	return &AdvancedTagParser{
		cache:   cache,
		fileSet: token.NewFileSet(),
		insights: &SemanticInsights{
			Patterns:        make(map[string]*CodePattern),
			Recommendations: []Recommendation{},
			Metrics: &ProjectMetrics{
				TagDistribution: make(map[TagType]int),
				ModuleMetrics:   make(map[string]*ModuleMetrics),
			},
		},
		depGraph: &DependencyGraph{
			Nodes:  make(map[string]*DependencyNode),
			Edges:  []*DependencyEdge{},
			Cycles: [][]string{},
			Levels: make(map[string]int),
		},
		config: config,
	}
}

// NewMemoryCache creates a new memory cache
func NewMemoryCache(ttl time.Duration) *MemoryCache {
	return &MemoryCache{
		data:   make(map[string]*CacheEntry),
		maxTTL: ttl,
	}
}

// Get retrieves data from cache
func (c *MemoryCache) Get(key string) ([]byte, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return nil, false
	}

	// Check TTL
	if time.Since(entry.CreatedAt) > entry.TTL {
		delete(c.data, key)
		return nil, false
	}

	return entry.Data, true
}

// Set stores data in cache
func (c *MemoryCache) Set(key string, value []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = &CacheEntry{
		Data:      value,
		CreatedAt: time.Now(),
		TTL:       c.maxTTL,
	}

	return nil
}

// Delete removes data from cache
func (c *MemoryCache) Delete(key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.data, key)
	return nil
}

// Clear removes all data from cache
func (c *MemoryCache) Clear() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data = make(map[string]*CacheEntry)
	return nil
}

// AnalyzeProjectAdvanced performs advanced project analysis
func (p *AdvancedTagParser) AnalyzeProjectAdvanced(ctx context.Context, projectPath string) (*ProjectAnalysis, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Check cache first
	cacheKey := p.generateCacheKey(projectPath)
	if cached, found := p.cache.Get(cacheKey); found {
		var analysis ProjectAnalysis
		if err := json.Unmarshal(cached, &analysis); err == nil {
			return &analysis, nil
		}
	}

	// Start fresh analysis
	analysis := &ProjectAnalysis{
		ProjectPath: projectPath,
		Modules:     make(map[string]*Module),
		LastScanned: time.Now(),
	}

	// Reset insights and dependency graph
	p.insights = &SemanticInsights{
		Patterns:        make(map[string]*CodePattern),
		Recommendations: []Recommendation{},
		Metrics: &ProjectMetrics{
			TagDistribution: make(map[TagType]int),
			ModuleMetrics:   make(map[string]*ModuleMetrics),
		},
	}
	p.depGraph = &DependencyGraph{
		Nodes:  make(map[string]*DependencyNode),
		Edges:  []*DependencyEdge{},
		Cycles: [][]string{},
		Levels: make(map[string]int),
	}

	// Walk project files
	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip non-Go files and ignored patterns
		if !strings.HasSuffix(path, ".go") || p.shouldIgnoreFile(path) {
			return nil
		}

		// Check file size
		if p.config.MaxFileSize > 0 && info.Size() > p.config.MaxFileSize {
			return nil
		}

		// Analyze file
		fileAnalysis, err := p.analyzeFileAdvanced(path)
		if err != nil {
			return err
		}

		// Merge results
		analysis.Tags = append(analysis.Tags, fileAnalysis.Tags...)
		p.mergeFileAnalysis(analysis, fileAnalysis)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to analyze project: %w", err)
	}

	// Perform semantic analysis
	if p.config.SemanticAnalysis {
		p.performSemanticAnalysis(analysis)
	}

	// Build dependency graph
	p.buildDependencyGraph(analysis)

	// Detect circular dependencies
	if p.config.CircularDetection {
		p.detectCircularDependencies()
	}

	// Generate recommendations
	p.generateRecommendations()

	// Calculate metrics
	p.calculateMetrics(analysis)

	// Cache results
	if data, err := json.Marshal(analysis); err == nil {
		p.cache.Set(cacheKey, data)
	}

	return analysis, nil
}
