package scanner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"regexp"
	"strings"
	"sync"

	"github.com/zricethezav/gitleaks/v8/detect"
)

type Severity string

const (
	SeverityHigh   Severity = "high"
	SeverityMedium Severity = "medium"
)

type Finding struct {
	Rule        string   `json:"rule"`
	Description string   `json:"description"`
	Severity    Severity `json:"severity"`
	File        string   `json:"file"`
	Line        int      `json:"line"`
	Fingerprint string   `json:"fingerprint"`
}

type Candidate struct {
	File string
	Line int
	Text string
}

var assignment = regexp.MustCompile(`(?i)(?:api[_-]?key|secret|token|password|passwd|credential|private[_-]?key)\s*[:=]\s*["']?([^\s"';#,}]{12,})`)
var placeholder = regexp.MustCompile(`(?i)^(?:example|sample|placeholder|changeme|replace[_-]?me|your[_-].*|test|dummy|xxx+|\$\{.*\}|<.*>)$`)
var credentialURL = regexp.MustCompile(`(?i)\b(?:postgres(?:ql)?|mysql|mongodb(?:\+srv)?|redis)://[^\s/:]+:([^\s/@]+)@`)

type engine struct {
	detector *detect.Detector
	mu       sync.Mutex
}

var (
	defaultEngine     *engine
	defaultEngineErr  error
	defaultEngineOnce sync.Once
)

func Scan(candidates []Candidate) ([]Finding, error) {
	defaultEngineOnce.Do(func() {
		detector, err := detect.NewDetectorDefaultConfig()
		if err != nil {
			defaultEngineErr = fmt.Errorf("load Gitleaks rules: %w", err)
			return
		}
		defaultEngine = &engine{detector: detector}
	})
	if defaultEngineErr != nil {
		return nil, defaultEngineErr
	}
	return defaultEngine.scan(candidates), nil
}

func (e *engine) scan(candidates []Candidate) []Finding {
	e.mu.Lock()
	defer e.mu.Unlock()
	var findings []Finding
	seen := map[string]bool{}
	for _, c := range candidates {
		if strings.Contains(c.Text, "dotenvai:allow") {
			continue
		}
		providerFindings := e.detector.DetectContext(context.Background(), detect.Fragment{
			Raw:       c.Text,
			FilePath:  c.File,
			StartLine: c.Line,
		})
		for _, provider := range providerFindings {
			severity := SeverityHigh
			if provider.RuleID == "generic-api-key" {
				severity = SeverityMedium
			}
			f := finding(provider.RuleID, provider.Description, severity, c, provider.Secret)
			if !seen[f.Fingerprint] {
				findings, seen[f.Fingerprint] = append(findings, f), mark(seen, f.Fingerprint)
			}
		}
		matched := len(providerFindings) > 0
		if match := credentialURL.FindStringSubmatch(c.Text); !matched && len(match) == 2 && !placeholder.MatchString(match[1]) {
			f := finding("credential-url", "Credential embedded in a database connection URL", SeverityHigh, c, match[1])
			findings, seen[f.Fingerprint] = append(findings, f), mark(seen, f.Fingerprint)
			matched = true
		}
		if match := assignment.FindStringSubmatch(c.Text); !matched && len(match) == 2 && suspicious(match[1]) {
			f := finding("generic-secret", "Likely secret assigned to a sensitive name", SeverityMedium, c, match[1])
			if !seen[f.Fingerprint] {
				findings, seen[f.Fingerprint] = append(findings, f), mark(seen, f.Fingerprint)
			}
		}
	}
	return findings
}

func suspicious(value string) bool {
	value = strings.Trim(value, "\"'")
	return len(value) >= 12 && !placeholder.MatchString(value) && (entropy(value) >= 3.0 || len(value) >= 24)
}

func entropy(s string) float64 {
	counts := map[rune]int{}
	for _, r := range s {
		counts[r]++
	}
	var result float64
	for _, n := range counts {
		p := float64(n) / float64(len([]rune(s)))
		result -= p * math.Log2(p)
	}
	return result
}

func finding(id, description string, severity Severity, c Candidate, secret string) Finding {
	h := sha256.Sum256([]byte(id + "\x00" + c.File + "\x00" + secret))
	return Finding{Rule: id, Description: description, Severity: severity, File: c.File, Line: c.Line, Fingerprint: hex.EncodeToString(h[:8])}
}

func mark(m map[string]bool, key string) bool { m[key] = true; return true }
