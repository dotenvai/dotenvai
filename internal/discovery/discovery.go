package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	WellKnownPath = "/.well-known/agent-skills.json"
	maxDocument   = 1 << 20
)

var (
	validName    = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	validVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?$`)
	validSHA256  = regexp.MustCompile(`^[a-f0-9]{64}$`)
)

type Document struct {
	SchemaVersion string    `json:"schema_version"`
	Publisher     Publisher `json:"publisher"`
	Skills        []Skill   `json:"skills"`
}

type Publisher struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Skill struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Version       string   `json:"version"`
	Source        string   `json:"source"`
	Archive       string   `json:"archive,omitempty"`
	SHA256        string   `json:"sha256,omitempty"`
	License       string   `json:"license,omitempty"`
	Compatibility []string `json:"compatibility,omitempty"`
}

func ReadFile(path string) (Document, error) {
	file, err := os.Open(path)
	if err != nil {
		return Document{}, err
	}
	defer file.Close()
	return Decode(file)
}

func Decode(reader io.Reader) (Document, error) {
	contents, err := io.ReadAll(io.LimitReader(reader, maxDocument+1))
	if err != nil {
		return Document{}, fmt.Errorf("read discovery document: %w", err)
	}
	if len(contents) > maxDocument {
		return Document{}, fmt.Errorf("discovery document exceeds 1 MiB")
	}
	var document Document
	decoder := json.NewDecoder(strings.NewReader(string(contents)))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&document); err != nil {
		return Document{}, fmt.Errorf("invalid discovery document: %w", err)
	}
	if err = decoder.Decode(&struct{}{}); err != io.EOF {
		return Document{}, fmt.Errorf("invalid discovery document: trailing content")
	}
	if err := document.Validate(); err != nil {
		return Document{}, err
	}
	return document, nil
}

func Fetch(ctx context.Context, target string) (Document, string, error) {
	discoveryURL, err := ResolveURL(target)
	if err != nil {
		return Document{}, "", err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return Document{}, "", err
	}
	request.Header.Set("Accept", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return Document{}, discoveryURL, err
	}
	defer response.Body.Close()
	if response.Request.URL.Scheme != "https" {
		return Document{}, discoveryURL, fmt.Errorf("discovery redirect resolved to a non-HTTPS URL")
	}
	if response.StatusCode != http.StatusOK {
		return Document{}, discoveryURL, fmt.Errorf("discovery endpoint returned %s", response.Status)
	}
	document, err := Decode(response.Body)
	return document, discoveryURL, err
}

func ResolveURL(target string) (string, error) {
	if !strings.Contains(target, "://") {
		target = "https://" + target + WellKnownPath
	}
	parsed, err := url.Parse(target)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return "", fmt.Errorf("discovery target must be an HTTPS URL or domain")
	}
	if parsed.User != nil || parsed.Fragment != "" {
		return "", fmt.Errorf("discovery URL must not contain credentials or a fragment")
	}
	return parsed.String(), nil
}

func (document Document) Validate() error {
	if document.SchemaVersion != "0.1" {
		return fmt.Errorf("unsupported schema_version %q (expected 0.1)", document.SchemaVersion)
	}
	if strings.TrimSpace(document.Publisher.Name) == "" {
		return fmt.Errorf("publisher.name is required")
	}
	if err := requireHTTPS("publisher.url", document.Publisher.URL); err != nil {
		return err
	}
	if len(document.Skills) == 0 {
		return fmt.Errorf("at least one skill is required")
	}
	seen := map[string]bool{}
	for i, skill := range document.Skills {
		prefix := fmt.Sprintf("skills[%d]", i)
		if !validName.MatchString(skill.Name) || len(skill.Name) > 64 {
			return fmt.Errorf("%s.name is invalid", prefix)
		}
		if seen[skill.Name] {
			return fmt.Errorf("duplicate skill name %q", skill.Name)
		}
		seen[skill.Name] = true
		if len(skill.Description) == 0 || len(skill.Description) > 1024 {
			return fmt.Errorf("%s.description must be 1-1024 characters", prefix)
		}
		if !validVersion.MatchString(skill.Version) {
			return fmt.Errorf("%s.version must be semantic versioning without a leading v", prefix)
		}
		if err := requireHTTPS(prefix+".source", skill.Source); err != nil {
			return err
		}
		if skill.Archive != "" {
			if err := requireHTTPS(prefix+".archive", skill.Archive); err != nil {
				return err
			}
			if !validSHA256.MatchString(skill.SHA256) {
				return fmt.Errorf("%s.sha256 must be 64 lowercase hex characters when archive is set", prefix)
			}
		} else if skill.SHA256 != "" {
			return fmt.Errorf("%s.archive is required when sha256 is set", prefix)
		}
	}
	return nil
}

func requireHTTPS(field, value string) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return fmt.Errorf("%s must be an absolute HTTPS URL", field)
	}
	return nil
}
