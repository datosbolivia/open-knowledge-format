package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Severity string

const (
	SeverityError   Severity = "ERROR"
	SeverityWarning Severity = "WARNING"
	SeverityInfo    Severity = "INFO"
)

type ValidationIssue struct {
	Code     string   `json:"code"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	Path     string   `json:"path,omitempty"`
	Line     int      `json:"line,omitempty"`
}

type ValidationReport struct {
	Valid         bool              `json:"valid"`
	Target        string            `json:"target"`
	TotalErrors   int               `json:"total_errors"`
	TotalWarnings int               `json:"total_warnings"`
	Issues        []ValidationIssue `json:"issues"`
}

func (r *ValidationReport) AddIssue(code string, sev Severity, msg string, path string, line int) {
	r.Issues = append(r.Issues, ValidationIssue{
		Code:     code,
		Severity: sev,
		Message:  msg,
		Path:     path,
		Line:     line,
	})
	if sev == SeverityError {
		r.TotalErrors++
		r.Valid = false
	} else if sev == SeverityWarning {
		r.TotalWarnings++
	}
}

var (
	sha256Regex       = regexp.MustCompile(`^sha256:[a-fA-F0-9]{64}$`)
	triadRegex        = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+[:/][a-zA-Z0-9_\-\. ]+([:/][a-zA-Z0-9_\-\. ]+)?$`)
	actorRegex        = regexp.MustCompile(`^([^<]+<[^>]+@[\w\.-]+>|https?://[^\s]+|agent:[a-zA-Z0-9_\-\.]+|user:[a-zA-Z0-9_\-\.]+|process:[a-zA-Z0-9_\-\.]+|human:[a-zA-Z0-9_\-\.]+)$`)
	isoDurationRegex  = regexp.MustCompile(`^P(?:\d+Y)?(?:\d+M)?(?:\d+W)?(?:\d+D)?(?:T(?:\d+H)?(?:\d+M)?(?:\d+S)?)?$`)
	markdownLinkRegex = regexp.MustCompile(`\[[^\]]*\]\(([^)]+\.md)\)`)

	validPolicies = map[string]bool{
		"allow_all":      true,
		"zero_microdata": true,
		"deny":           true,
	}
	validContractTypes = map[string]bool{
		"datapackage": true,
		"dcat":        true,
		"openapi":     true,
		"cwl":         true,
		"cli_tool":    true,
	}
	validStatusMap = map[string]bool{
		"draft":      true,
		"proposed":   true,
		"active":     true,
		"deprecated": true,
		"retired":    true,
		"superseded": true,
		"stable":     true,
		"verified":   true,
		"refuted":    true,
	}
)

type ConceptFrontmatter struct {
	Type         string                 `yaml:"type"`
	Title        string                 `yaml:"title"`
	Description  string                 `yaml:"description"`
	Resource     string                 `yaml:"resource"`
	Tags         []string               `yaml:"tags"`
	Timestamp    string                 `yaml:"timestamp"`
	Status       string                 `yaml:"status"`
	SupersededBy string                 `yaml:"superseded_by"`
	Generated    *GeneratedSpec         `yaml:"generated"`
	Verified     []VerificationRecord   `yaml:"verified"`
	Freshness    *FreshnessSpec         `yaml:"freshness"`
	SKOS         []string               `yaml:"skos"`
	Dimensions   []string               `yaml:"dimensions"`
	Contracts    []ContractSpec         `yaml:"contracts"`
	Lineage      *LineageSpec           `yaml:"lineage"`
	Policy       string                 `yaml:"policy"`
	Extra        map[string]interface{} `yaml:",inline"`
}

type ContractSpec struct {
	Type string `yaml:"type" json:"type"`
	Path string `yaml:"path" json:"path"`
	Ref  string `yaml:"ref" json:"ref"`
}

type LineageSpec struct {
	Source          []map[string]interface{} `yaml:"source" json:"source"`
	Transformations []TransformationSpec     `yaml:"transformations" json:"transformations"`
	Version         string                   `yaml:"version" json:"version"`
	UpdatedAt       string                   `yaml:"updated_at" json:"updated_at"`
}

type TransformationSpec struct {
	Step string `yaml:"step" json:"step"`
	Tool string `yaml:"tool" json:"tool"`
	Hash string `yaml:"hash" json:"hash"`
}

func parseFlexibleTimestamp(ts string) bool {
	clean := strings.TrimSpace(ts)
	if clean == "" {
		return false
	}
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if _, err := time.Parse(f, clean); err == nil {
			return true
		}
	}
	return false
}

// ValidateConceptContent validates an in-memory concept string against OKF/ODKF rules.
func ValidateConceptContent(content string, filePath string) *ValidationReport {
	report := &ValidationReport{
		Valid:  true,
		Target: filePath,
		Issues: make([]ValidationIssue, 0),
	}

	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		report.AddIssue("PE1-SYNTAX", SeverityError, "Document must begin with YAML frontmatter delimiter '---'", filePath, 1)
		return report
	}

	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		report.AddIssue("PE1-SYNTAX", SeverityError, "Unclosed YAML frontmatter: missing ending delimiter '---'", filePath, len(lines))
		return report
	}

	frontmatterText := strings.Join(lines[1:endIdx], "\n")
	var fm ConceptFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatterText), &fm); err != nil {
		report.AddIssue("RULE-01-YAML", SeverityError, fmt.Sprintf("YAML frontmatter parse error: %v", err), filePath, 1)
		return report
	}

	// RULE-02 / PE2: Mandatory type
	if strings.TrimSpace(fm.Type) == "" {
		report.AddIssue("PE2-TYPE-REQUIRED", SeverityError, "Field 'type' is required in concept frontmatter", filePath, 1)
	}

	// RULE-03: Semantic Anchors
	normType := strings.ToLower(strings.TrimSpace(fm.Type))
	if normType == "dataset" && len(fm.SKOS) == 0 {
		report.AddIssue("RULE-03-SKOS-MISSING", SeverityWarning, "Dataset concept should define at least one ontological anchor in 'skos'", filePath, 1)
	}
	for _, s := range fm.SKOS {
		u, err := url.ParseRequestURI(s)
		if err != nil || u.Scheme == "" || u.Host == "" {
			report.AddIssue("RULE-03-SKOS-INVALID", SeverityError, fmt.Sprintf("Invalid URI in skos anchor: '%s'", s), filePath, 1)
		}
	}

	// RULE-04 / PE4: Syntactic Contracts
	if normType == "dataset" && len(fm.Contracts) == 0 {
		report.AddIssue("PE4-CONTRACT-MISSING", SeverityWarning, "Dataset concept defines no syntactic contract in 'contracts'", filePath, 1)
	}
	for idx, c := range fm.Contracts {
		if strings.TrimSpace(c.Type) == "" {
			report.AddIssue("PE4-CONTRACT-TYPE", SeverityError, fmt.Sprintf("Contract[%d] is missing 'type'", idx), filePath, 1)
		} else if !validContractTypes[strings.ToLower(strings.TrimSpace(c.Type))] {
			report.AddIssue("PE4-CONTRACT-UNKNOWN", SeverityWarning, fmt.Sprintf("Contract[%d] has non-standard type '%s'", idx, c.Type), filePath, 1)
		}

		targetRef := strings.TrimSpace(c.Path)
		if targetRef == "" {
			targetRef = strings.TrimSpace(c.Ref)
		}
		if targetRef == "" {
			report.AddIssue("PE4-CONTRACT-PATH", SeverityError, fmt.Sprintf("Contract[%d] is missing 'path' or 'ref'", idx), filePath, 1)
		} else if !strings.HasPrefix(targetRef, "http://") && !strings.HasPrefix(targetRef, "https://") && filePath != "" {
			baseDir := filepath.Dir(filePath)
			targetFile := filepath.Join(baseDir, targetRef)
			if _, err := os.Stat(targetFile); os.IsNotExist(err) {
				report.AddIssue("PE4-CONTRACT-LOCAL-NOT-FOUND", SeverityWarning, fmt.Sprintf("Local contract path '%s' not found at '%s'", targetRef, targetFile), filePath, 1)
			}
		}
	}

	// RULE-05 / PE5: Cryptographic Hash Integrity
	if fm.Lineage != nil {
		for idx, t := range fm.Lineage.Transformations {
			if strings.TrimSpace(t.Hash) != "" {
				if !sha256Regex.MatchString(strings.TrimSpace(t.Hash)) {
					report.AddIssue("PE5-HASH-FORMAT", SeverityError, fmt.Sprintf("Transformation[%d] hash '%s' does not match format 'sha256:<hex64>'", idx, t.Hash), filePath, 1)
				} else if t.Tool != "" && filePath != "" {
					baseDir := filepath.Dir(filePath)
					toolFile := filepath.Join(baseDir, t.Tool)
					if fi, err := os.Stat(toolFile); err == nil && !fi.IsDir() {
						computed, err := computeFileSHA256(toolFile)
						if err == nil {
							expected := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(t.Hash), "sha256:"))
							if computed != expected {
								report.AddIssue("PE5-HASH-MISMATCH", SeverityError, fmt.Sprintf("Computed SHA-256 ('%s') of tool '%s' does not match declared hash ('%s')", computed, t.Tool, expected), filePath, 1)
							}
						}
					}
				}
			}
		}
	}

	// RULE-06: Policy Enum
	if strings.TrimSpace(fm.Policy) != "" {
		pNorm := strings.ToLower(strings.TrimSpace(fm.Policy))
		if !validPolicies[pNorm] {
			report.AddIssue("RULE-06-POLICY-INVALID", SeverityError, fmt.Sprintf("Policy '%s' is invalid. Must be allow_all, zero_microdata, or deny", fm.Policy), filePath, 1)
		}
	}

	// RULE-09: Actor Format (§7)
	if fm.Generated != nil && strings.TrimSpace(fm.Generated.By) != "" {
		if !actorRegex.MatchString(strings.TrimSpace(fm.Generated.By)) {
			report.AddIssue("RULE-09-ACTOR-FORMAT", SeverityWarning, fmt.Sprintf("generated.by '%s' does not match actor format ('Name <email>', URL, or actor URI prefix)", fm.Generated.By), filePath, 1)
		}
	}
	for idx, v := range fm.Verified {
		if strings.TrimSpace(v.By) != "" && !actorRegex.MatchString(strings.TrimSpace(v.By)) {
			report.AddIssue("RULE-09-ACTOR-FORMAT", SeverityWarning, fmt.Sprintf("verified[%d].by '%s' does not match actor format ('Name <email>', URL, or actor URI prefix)", idx, v.By), filePath, 1)
		}
	}

	// RULE-10: Status Enum & Superseded (§5.3)
	if strings.TrimSpace(fm.Status) != "" {
		stNorm := strings.ToLower(strings.TrimSpace(fm.Status))
		if !validStatusMap[stNorm] {
			report.AddIssue("RULE-10-STATUS-UNKNOWN", SeverityWarning, fmt.Sprintf("status '%s' is not in canonical enum (draft, proposed, active, deprecated, retired, superseded)", fm.Status), filePath, 1)
		}
		if stNorm == "superseded" && strings.TrimSpace(fm.SupersededBy) == "" {
			report.AddIssue("RULE-10-SUPERSEDED-TARGET", SeverityError, "Concept marked as 'superseded' MUST define 'superseded_by' with target path or URI", filePath, 1)
		}
	}

	// RULE-11: ISO 8601 Timestamps & Durations (§5.1, §5.4)
	if fm.Freshness != nil {
		if strings.TrimSpace(fm.Freshness.Cadence) != "" {
			if !isoDurationRegex.MatchString(strings.TrimSpace(fm.Freshness.Cadence)) {
				report.AddIssue("RULE-11-CADENCE-SYNTAX", SeverityWarning, fmt.Sprintf("freshness.cadence '%s' must follow ISO 8601 duration format (e.g. P1D, P1W, P1M)", fm.Freshness.Cadence), filePath, 1)
			}
		}
		if strings.TrimSpace(fm.Freshness.Expires) != "" && !parseFlexibleTimestamp(fm.Freshness.Expires) {
			report.AddIssue("RULE-11-EXPIRES-SYNTAX", SeverityWarning, fmt.Sprintf("freshness.expires '%s' must be a valid ISO 8601 date or timestamp", fm.Freshness.Expires), filePath, 1)
		}
	}
	if fm.Generated != nil && strings.TrimSpace(fm.Generated.At) != "" && !parseFlexibleTimestamp(fm.Generated.At) {
		report.AddIssue("RULE-11-GENERATED-AT-SYNTAX", SeverityWarning, fmt.Sprintf("generated.at '%s' must be a valid ISO 8601 timestamp", fm.Generated.At), filePath, 1)
	}
	for idx, v := range fm.Verified {
		if strings.TrimSpace(v.At) != "" && !parseFlexibleTimestamp(v.At) {
			report.AddIssue("RULE-11-VERIFIED-AT-SYNTAX", SeverityWarning, fmt.Sprintf("verified[%d].at '%s' must be a valid ISO 8601 timestamp", idx, v.At), filePath, 1)
		}
	}

	// Check legacy Timestamp format if present
	if strings.TrimSpace(fm.Timestamp) != "" && !parseFlexibleTimestamp(fm.Timestamp) {
		report.AddIssue("RULE-01-TIMESTAMP", SeverityWarning, fmt.Sprintf("Timestamp '%s' does not follow ISO 8601", fm.Timestamp), filePath, 1)
	}

	return report
}

// ValidateTriad validates canonical triad addressing syntax.
func ValidateTriad(triad string) *ValidationReport {
	report := &ValidationReport{
		Valid:  true,
		Target: triad,
		Issues: make([]ValidationIssue, 0),
	}

	clean := strings.Trim(strings.TrimSpace(triad), `"'`)
	if !triadRegex.MatchString(clean) {
		report.AddIssue("RULE-08-TRIAD-SYNTAX", SeverityError, fmt.Sprintf("Triad '%s' does not match canonical pattern 'catalog:dataset:resource' or 'catalog/dataset/resource'", clean), "", 0)
	}

	return report
}

// ValidateBundle recursively validates an OKF/ODKF bundle directory, checking all concepts and detecting cyclic references.
func ValidateBundle(bundleDir string) (*ValidationReport, error) {
	report := &ValidationReport{
		Valid:  true,
		Target: bundleDir,
		Issues: make([]ValidationIssue, 0),
	}

	info, err := os.Stat(bundleDir)
	if err != nil {
		return nil, fmt.Errorf("bundle directory error: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("target '%s' is not a directory", bundleDir)
	}

	conceptContents := make(map[string]string)
	conceptCount := 0

	err = filepath.Walk(bundleDir, func(path string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if fi.IsDir() {
			return nil
		}

		baseName := fi.Name()
		if !strings.HasSuffix(baseName, ".md") {
			return nil
		}

		contentBytes, err := os.ReadFile(path)
		if err != nil {
			report.AddIssue("IO-READ-ERROR", SeverityError, fmt.Sprintf("Could not read file: %v", err), path, 0)
			return nil
		}
		content := string(contentBytes)

		// Check reserved files: index.md and log.md
		if baseName == "index.md" || baseName == "log.md" {
			if strings.HasPrefix(strings.TrimSpace(content), "---") {
				var fm ConceptFrontmatter
				parts := strings.Split(content, "---")
				if len(parts) >= 3 {
					_ = yaml.Unmarshal([]byte(parts[1]), &fm)
					if fm.Type == "dataset" || fm.Type == "service" {
						report.AddIssue("PE3-RESERVED-FILE", SeverityWarning, fmt.Sprintf("Reserved file '%s' defines a concept type '%s'", baseName, fm.Type), path, 1)
					}
				}
			}
			return nil
		}

		conceptCount++
		conceptContents[path] = content

		docReport := ValidateConceptContent(content, path)
		if !docReport.Valid {
			report.Valid = false
		}
		report.TotalErrors += docReport.TotalErrors
		report.TotalWarnings += docReport.TotalWarnings
		report.Issues = append(report.Issues, docReport.Issues...)

		return nil
	})

	if err != nil {
		return nil, err
	}

	if conceptCount == 0 {
		report.AddIssue("RULE-07-EMPTY-BUNDLE", SeverityError, "Bundle directory contains no concept markdown (.md) documents", bundleDir, 0)
		return report, nil
	}

	// RULE-12: Directed Graph Cycle Detection across concepts in bundle
	graph := make(map[string][]string)
	for path, content := range conceptContents {
		dir := filepath.Dir(path)
		matches := markdownLinkRegex.FindAllStringSubmatch(content, -1)
		for _, m := range matches {
			if len(m) >= 2 {
				targetRef := m[1]
				if strings.HasPrefix(targetRef, "http://") || strings.HasPrefix(targetRef, "https://") || strings.HasPrefix(targetRef, "#") {
					continue
				}
				cleanTarget := filepath.Clean(filepath.Join(dir, targetRef))
				if _, exists := conceptContents[cleanTarget]; exists && cleanTarget != path {
					graph[path] = append(graph[path], cleanTarget)
				}
			}
		}
	}

	// DFS 3-color cycle detection: 0 = White (unvisited), 1 = Gray (visiting), 2 = Black (visited)
	color := make(map[string]int)
	var hasCycle func(node string, pathStack []string) bool
	hasCycle = func(node string, pathStack []string) bool {
		color[node] = 1
		pathStack = append(pathStack, node)

		for _, neighbor := range graph[node] {
			if color[neighbor] == 1 {
				cycleStr := ""
				for _, p := range pathStack {
					cycleStr += filepath.Base(p) + " -> "
				}
				cycleStr += filepath.Base(neighbor)
				report.AddIssue("RULE-12-CIRCULAR-REFERENCE", SeverityError, fmt.Sprintf("Circular dependency detected between concepts: %s", cycleStr), node, 1)
				return true
			}
			if color[neighbor] == 0 {
				if hasCycle(neighbor, pathStack) {
					return true
				}
			}
		}

		color[node] = 2
		return false
	}

	for node := range conceptContents {
		if color[node] == 0 {
			hasCycle(node, nil)
		}
	}

	return report, nil
}

func computeFileSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ToJSON serializes report to JSON string.
func (r *ValidationReport) ToJSON() (string, error) {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
