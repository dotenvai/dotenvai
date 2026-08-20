package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var validName = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Metadata struct {
	Name          string            `yaml:"name" json:"name"`
	Description   string            `yaml:"description" json:"description"`
	License       string            `yaml:"license,omitempty" json:"license,omitempty"`
	Compatibility string            `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`
	Metadata      map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

func Validate(path string) (Metadata, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Metadata{}, err
	}
	dir := path
	if info.IsDir() {
		path = filepath.Join(path, "SKILL.md")
	} else {
		dir = filepath.Dir(path)
		if filepath.Base(path) != "SKILL.md" {
			return Metadata{}, fmt.Errorf("skill entrypoint must be named SKILL.md")
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Metadata{}, err
	}
	frontmatter, err := parseFrontmatter(string(data))
	if err != nil {
		return Metadata{}, err
	}
	var metadata Metadata
	if err := yaml.Unmarshal([]byte(frontmatter), &metadata); err != nil {
		return Metadata{}, fmt.Errorf("invalid YAML frontmatter: %w", err)
	}
	if !validName.MatchString(metadata.Name) || len(metadata.Name) > 64 {
		return Metadata{}, fmt.Errorf("name must be 1-64 lowercase letters, numbers, or single hyphens")
	}
	if metadata.Name != filepath.Base(dir) {
		return Metadata{}, fmt.Errorf("name %q must match directory %q", metadata.Name, filepath.Base(dir))
	}
	if n := len(metadata.Description); n == 0 || n > 1024 {
		return Metadata{}, fmt.Errorf("description must be 1-1024 characters")
	}
	if len(metadata.Compatibility) > 500 {
		return Metadata{}, fmt.Errorf("compatibility must not exceed 500 characters")
	}
	return metadata, nil
}

func parseFrontmatter(contents string) (string, error) {
	contents = strings.ReplaceAll(contents, "\r\n", "\n")
	if !strings.HasPrefix(contents, "---\n") {
		return "", fmt.Errorf("SKILL.md must begin with YAML frontmatter")
	}
	end := strings.Index(contents[4:], "\n---\n")
	if end < 0 {
		return "", fmt.Errorf("SKILL.md frontmatter is not closed with ---")
	}
	return contents[4 : 4+end], nil
}
