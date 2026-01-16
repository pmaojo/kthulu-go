package skills

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Skill represents an agent skill
type Skill struct {
	Name        string
	Description string
	Path        string // Absolute path to SKILL.md
}

// SkillMetadata matches the YAML frontmatter
type SkillMetadata struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// Manager manages loaded skills
type Manager struct {
	skills []Skill
}

// NewManager creates a new skill manager
func NewManager() *Manager {
	return &Manager{
		skills: make([]Skill, 0),
	}
}

// All returns all loaded skills
func (m *Manager) All() []Skill {
	return m.skills
}

// Get returns a skill by name
func (m *Manager) Get(name string) (Skill, bool) {
	for _, s := range m.skills {
		if strings.EqualFold(s.Name, name) {
			return s, true
		}
	}
	return Skill{}, false
}

// LoadSkills searches for SKILL.md files in the given root directories
// and recursively loads them. Typical root is ".kthulu/skills".
func (m *Manager) LoadSkills(rootDirs ...string) error {
	for _, root := range rootDirs {
		if _, err := os.Stat(root); os.IsNotExist(err) {
			continue
		}

		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.EqualFold(d.Name(), "SKILL.md") {
				if skill, err := ParseSkill(path); err == nil {
					m.skills = append(m.skills, skill)
				}
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("error walking skills directory %s: %w", root, err)
		}
	}
	return nil
}

// ParseSkill parses a SKILL.md file and returns a Skill struct
func ParseSkill(path string) (Skill, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Skill{}, err
	}

	parts := strings.SplitN(string(content), "---", 3)
	if len(parts) < 3 {
		return Skill{}, fmt.Errorf("invalid SKILL.md format: missing frontmatter")
	}

	// Frontmatter is parts[1] (parts[0] is empty string before first ---)
	var meta SkillMetadata
	if err := yaml.Unmarshal([]byte(parts[1]), &meta); err != nil {
		return Skill{}, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// If name is missing, directory name fallback or error?
	// The user request emphasizes "name and description" in frontmatter.
	if meta.Name == "" {
		return Skill{}, fmt.Errorf("skill name required in frontmatter")
	}

	return Skill{
		Name:        meta.Name,
		Description: meta.Description,
		Path:        path,
	}, nil
}
