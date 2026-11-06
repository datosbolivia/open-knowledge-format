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
	triadRegex        = regexp.MustCompile(`^([a-zA-Z0-9_\-\.]+[:/][a-zA-Z0-9_\-\. ]+([:/][a-zA-Z0-9_\-\. ]+)?|[a-zA-Z0-9_\-\.]+\.[a-zA-Z0-9_\-\.]+@[a-zA-Z0-9_\-\.]+|okf://[a-zA-Z0-9_\-\./]+|datamesh://[a-zA-Z0-9_\-\./]+)(#[a-zA-Z0-9_\-\./]+)?$`)
	actorRegex        = regexp.MustCompile(`^([^<]+<[^>]+@[\w\.-]+>|https?://[^\s]+|[a-zA-Z0-9_\-\.]+/[a-zA-Z0-9_\-\.]+|(agent|user|process|human|team):[a-zA-Z0-9_\-\.@/]+)$`)
	isoDurationRegex  = regexp.MustCompile(`^P(?:\d+Y)?(?:\d+M)?(?:\d+W)?(?:\d+D)?(?:T(?:\d+H)?(?:\d+M)?(?:\d+S)?)?$`)
	markdownLinkRegex = regexp.MustCompile(`\[[^\]]*\]\(([^)]+\.md)\)`)

	validPolicies = map[string]bool{
		"allow_all":      true,
		"zero_microdata": true,
		"deny":           true,
	}
	validContractTypes = map[string]bool{
		"datapackage":  true,
		"frictionless": true,
		"dcat":         true,
		"openapi":      true,
		"asyncapi":     true,
		"cwl":          true,
		"cli_tool":     true,
		"json_schema":  true,
		"sdmx":         true,
		"ddi":          true,
		"iso19115":     true,
		"stac":         true,
		"graphql":      true,
		"flight_sql":   true,
		"odrl":         true,
		"opa":          true,
		"dpv":          true,
		"owl-cli":      true,
		"okf-cli":      true,
		"owl":          true,
		"rdf":          true,
		"ckan":         true,
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

type StringList []string

func (s *StringList) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.SequenceNode {
		var list []string
		if err := value.Decode(&list); err != nil {
			return err
		}
		*s = list
		return nil
	}
	if value.Kind == yaml.ScalarNode {
		var raw string
		if err := value.Decode(&raw); err != nil {
			return err
		}
		var list []string
		for _, part := range strings.Split(raw, ",") {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				list = append(list, trimmed)
			}
		}
		*s = list
		return nil
	}
	return nil
}

type VerificationList []VerificationRecord

func (v *VerificationList) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.SequenceNode {
		var list []VerificationRecord
		if err := value.Decode(&list); err != nil {
			return err
		}
		*v = list
		return nil
	}
	if value.Kind == yaml.MappingNode {
		var single VerificationRecord
		if err := value.Decode(&single); err != nil {
			return err
		}
		*v = []VerificationRecord{single}
		return nil
	}
	return nil
}

type RelationSpec struct {
	Rel    string `yaml:"rel" json:"rel"`       // derived_from, supersedes, foreign_key, references, in_catalog
	Target string `yaml:"target" json:"target"` // Path or URI
}

type ConceptFrontmatter struct {
	Type         string                 `yaml:"type"`
	Title        string                 `yaml:"title"`
	Description  string                 `yaml:"description"`
	Resource     string                 `yaml:"resource"`
	Tags         StringList             `yaml:"tags"`
	Timestamp    string                 `yaml:"timestamp"`
	Status       string                 `yaml:"status"`
	Supersedes   string                 `yaml:"supersedes"`
	SupersededBy string                 `yaml:"superseded_by"`
	StaleAfter   string                 `yaml:"stale_after"`
	Sources      []interface{}          `yaml:"sources"`
	Relations    []RelationSpec         `yaml:"relations"`
	Generated    *GeneratedSpec         `yaml:"generated"`
	Verified     VerificationList       `yaml:"verified"`
	Freshness    *FreshnessSpec         `yaml:"freshness"`
	SKOS         []string               `yaml:"skos"`
	ExactMatch   []string               `yaml:"exact_match"`
	CloseMatch   []string               `yaml:"close_match"`
	Broader      []string               `yaml:"broader"`
	Dimensions   []string               `yaml:"dimensions"`
	Contracts    []ContractSpec         `yaml:"contracts"`
	Lineage      *LineageSpec           `yaml:"lineage"`
	Policy       interface{}            `yaml:"policy"`
	Policies     []PolicySpec           `yaml:"policies"`
	Extra        map[string]interface{} `yaml:",inline"`
}

type ContractSpec struct {
	Type       string                 `yaml:"type" json:"type"`
	Path       string                 `yaml:"path" json:"path"`
	Ref        string                 `yaml:"ref" json:"ref"`
	Binary     string                 `yaml:"binary,omitempty" json:"binary,omitempty"`
	Subcommand string                 `yaml:"subcommand,omitempty" json:"subcommand,omitempty"`
	Inputs     []CLIInputParam        `yaml:"inputs,omitempty" json:"inputs,omitempty"`
	Outputs    *CLIOutputSpec         `yaml:"outputs,omitempty" json:"outputs,omitempty"`
	Inline     map[string]interface{} `yaml:"inline" json:"inline"`
}

type LineageSpec struct {
	Source          []interface{}        `yaml:"source" json:"source"`
	Sources         []interface{}        `yaml:"sources" json:"sources"`
	Transformations []TransformationSpec `yaml:"transformations" json:"transformations"`
	Version         string               `yaml:"version" json:"version"`
	UpdatedAt       string               `yaml:"updated_at" json:"updated_at"`
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

	// RULE-04 / PE4: Syntactic & Meta-Specification Contracts
	if normType == "dataset" && len(fm.Contracts) == 0 {
		report.AddIssue("PE4-CONTRACT-MISSING", SeverityWarning, "Dataset concept defines no syntactic contract in 'contracts'", filePath, 1)
	}
	for idx, c := range fm.Contracts {
		if strings.TrimSpace(c.Type) == "" {
			report.AddIssue("PE4-CONTRACT-TYPE", SeverityError, fmt.Sprintf("Contract[%d] is missing 'type'", idx), filePath, 1)
		} else if !validContractTypes[strings.ToLower(strings.TrimSpace(c.Type))] {
			report.AddIssue("PE4-CONTRACT-UNKNOWN", SeverityInfo, fmt.Sprintf("Contract[%d] uses custom/extensible specification type '%s'", idx, c.Type), filePath, 1)
		}

		targetRef := strings.TrimSpace(c.Path)
		if targetRef == "" {
			targetRef = strings.TrimSpace(c.Ref)
		}
		if targetRef == "" && len(c.Inline) == 0 && strings.TrimSpace(c.Binary) == "" && len(c.Inputs) == 0 {
			report.AddIssue("PE4-CONTRACT-PATH", SeverityError, fmt.Sprintf("Contract[%d] must specify 'path', 'ref', 'inline', or 'binary'", idx), filePath, 1)
		} else if targetRef != "" && !strings.HasPrefix(targetRef, "http://") && !strings.HasPrefix(targetRef, "https://") && filePath != "" {
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

	// RULE-06: Policy Governance (Shorthand Profiles & External Meta-Specification Policies)
	validatePolicyEntry := func(p PolicySpec, label string) {
		if strings.TrimSpace(p.Profile) != "" {
			pNorm := strings.ToLower(strings.TrimSpace(p.Profile))
			if !validPolicies[pNorm] && !strings.Contains(pNorm, "://") && !strings.HasPrefix(pNorm, "urn:") {
				report.AddIssue("RULE-06-POLICY-INVALID", SeverityError, fmt.Sprintf("%s profile '%s' is invalid. Must be allow_all, zero_microdata, deny, or a policy URI", label, p.Profile), filePath, 1)
			}
		}
		targetRef := strings.TrimSpace(p.Path)
		if targetRef == "" {
			targetRef = strings.TrimSpace(p.Ref)
		}
		if strings.TrimSpace(p.Type) == "" && strings.TrimSpace(p.Profile) == "" && targetRef == "" && len(p.Rules) == 0 && len(p.Inline) == 0 {
			report.AddIssue("RULE-06-POLICY-EMPTY", SeverityError, fmt.Sprintf("%s must define 'type', 'profile', 'path', 'ref', 'rules', or 'inline'", label), filePath, 1)
		}
		if targetRef != "" && !strings.HasPrefix(targetRef, "http://") && !strings.HasPrefix(targetRef, "https://") && filePath != "" {
			baseDir := filepath.Dir(filePath)
			targetFile := filepath.Join(baseDir, targetRef)
			if _, err := os.Stat(targetFile); os.IsNotExist(err) {
				report.AddIssue("RULE-06-POLICY-LOCAL-NOT-FOUND", SeverityWarning, fmt.Sprintf("Local policy path '%s' not found at '%s'", targetRef, targetFile), filePath, 1)
			}
		}
	}

	if fm.Policy != nil {
		switch pv := fm.Policy.(type) {
		case string:
			trimmed := strings.TrimSpace(pv)
			if trimmed != "" {
				pNorm := strings.ToLower(trimmed)
				isURI := strings.Contains(pNorm, "://") || strings.HasPrefix(pNorm, "urn:") || strings.HasPrefix(pNorm, "{")
				if !validPolicies[pNorm] && !isURI {
					report.AddIssue("RULE-06-POLICY-INVALID", SeverityError, fmt.Sprintf("Policy '%s' is invalid. Must be allow_all, zero_microdata, deny, a policy URI, or a structured policy specification", pv), filePath, 1)
				}
			}
		case map[string]interface{}:
			rawBytes, err := yaml.Marshal(pv)
			if err == nil {
				var ps PolicySpec
				if err := yaml.Unmarshal(rawBytes, &ps); err == nil {
					validatePolicyEntry(ps, "policy")
				}
			}
		case []interface{}:
			rawBytes, err := yaml.Marshal(pv)
			if err == nil {
				var psList []PolicySpec
				if err := yaml.Unmarshal(rawBytes, &psList); err == nil {
					for i, ps := range psList {
						validatePolicyEntry(ps, fmt.Sprintf("policy[%d]", i))
					}
				}
			}
		}
	}
	for idx, ps := range fm.Policies {
		validatePolicyEntry(ps, fmt.Sprintf("policies[%d]", idx))
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

	// RULE-11: ISO 8601 Timestamps & Durations (§5.1, §5.4, §5.5)
	if strings.TrimSpace(fm.StaleAfter) != "" && !parseFlexibleTimestamp(fm.StaleAfter) {
		report.AddIssue("RULE-11-EXPIRES-SYNTAX", SeverityWarning, fmt.Sprintf("stale_after '%s' must be a valid ISO 8601 date or timestamp", fm.StaleAfter), filePath, 1)
	}
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
					// Support Root Concept Manifests (e.g. single-dataset ODKF nodes where root index.md carries type: dcat:Dataset / dcat:DataService / dcat:CatalogRecord)
					if baseName == "index.md" && strings.TrimSpace(fm.Type) != "" {
						conceptCount++
						conceptContents[path] = content
						docReport := ValidateConceptContent(content, path)
						if !docReport.Valid {
							report.Valid = false
						}
						report.TotalErrors += docReport.TotalErrors
						report.TotalWarnings += docReport.TotalWarnings
						report.Issues = append(report.Issues, docReport.Issues...)
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

	// RULE-12: Directed Graph Cycle Detection across concepts in bundle.
	// Distinguishes navigational body links (where cycles are valid per SPEC.md §14.5 RULE-07 -> SeverityWarning)
	// from strict acyclic DAG relations (supersedes, superseded_by, derived_from -> SeverityError).
	navGraph := make(map[string][]string)
	dagGraph := make(map[string][]string)
	for path, content := range conceptContents {
		dir := filepath.Dir(path)
		matches := markdownLinkRegex.FindAllStringSubmatch(content, -1)
		for _, m := range matches {
			if len(m) >= 2 {
				targetRef := m[1]
				if strings.HasPrefix(targetRef, "http://") || strings.HasPrefix(targetRef, "https://") || strings.HasPrefix(targetRef, "#") {
					continue
				}
				var cleanTarget string
				if strings.HasPrefix(targetRef, "/") {
					cleanTarget = filepath.Clean(filepath.Join(bundleDir, strings.TrimPrefix(targetRef, "/")))
				} else {
					cleanTarget = filepath.Clean(filepath.Join(dir, targetRef))
				}
				if _, exists := conceptContents[cleanTarget]; exists && cleanTarget != path {
					navGraph[path] = append(navGraph[path], cleanTarget)
				}
			}
		}

		// Extract strict DAG edges from frontmatter (supersedes, superseded_by, relations[rel=derived_from|supersedes|in_catalog])
		parts := strings.Split(content, "---")
		if len(parts) >= 3 {
			var fm ConceptFrontmatter
			if err := yaml.Unmarshal([]byte(parts[1]), &fm); err == nil {
				addDAGEdge := func(ref string) {
					ref = strings.TrimSpace(ref)
					if ref == "" || strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
						return
					}
					var targetPath string
					if strings.HasPrefix(ref, "/") {
						targetPath = filepath.Clean(filepath.Join(bundleDir, strings.TrimPrefix(ref, "/")))
					} else {
						targetPath = filepath.Clean(filepath.Join(dir, ref))
					}
					if _, exists := conceptContents[targetPath]; exists && targetPath != path {
						dagGraph[path] = append(dagGraph[path], targetPath)
					}
				}
				addDAGEdge(fm.Supersedes)
				addDAGEdge(fm.SupersededBy)
				for _, rel := range fm.Relations {
					rNorm := strings.ToLower(strings.TrimSpace(rel.Rel))
					if rNorm == "derived_from" || rNorm == "supersedes" || rNorm == "in_catalog" || rNorm == "dcat:incatalog" {
						addDAGEdge(rel.Target)
					}
				}
			}
		}
	}

	detectGraphCycles := func(g map[string][]string, sev Severity, code string, label string) {
		color := make(map[string]int)
		var hasCycle func(node string, pathStack []string) bool
		hasCycle = func(node string, pathStack []string) bool {
			color[node] = 1
			pathStack = append(pathStack, node)

			for _, neighbor := range g[node] {
				if color[neighbor] == 1 {
					cycleStr := ""
					for _, p := range pathStack {
						cycleStr += filepath.Base(p) + " -> "
					}
					cycleStr += filepath.Base(neighbor)
					report.AddIssue(code, sev, fmt.Sprintf("%s detected between concepts: %s", label, cycleStr), node, 1)
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
	}

	// Strict DAG cycle check (ERROR)
	detectGraphCycles(dagGraph, SeverityError, "RULE-12-DAG-CYCLE", "Acyclic DAG violation (lineage/supersession cycle)")
	// Navigational markdown link cycle check (WARNING per SPEC.md §14.5 RULE-07)
	detectGraphCycles(navGraph, SeverityWarning, "RULE-12-CIRCULAR-REFERENCE", "Circular dependency")

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
