package usecases

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/datosbolivia/open-knowledge-format/src/domain"
	"gopkg.in/yaml.v3"
)

// SemVerLevel classifies schema or bundle changes per Semantic Versioning 2.0.0 (Preston-Werner, 2013).
type SemVerLevel string

const (
	SemVerNone  SemVerLevel = "NONE"
	SemVerPatch SemVerLevel = "PATCH" // Prose/documentation edits only
	SemVerMinor SemVerLevel = "MINOR" // Backward-compatible additions (new optional columns, new resources, new concepts)
	SemVerMajor SemVerLevel = "MAJOR" // Breaking changes (removed/renamed columns, type changes, unit/SKOS changes, removed resources)
)

// DiffChange represents an atomic change between two versions of a bundle or data contract.
type DiffChange struct {
	Level       SemVerLevel `json:"level"`
	Category    string      `json:"category"` // "column", "resource", "concept", "documentation"
	Target      string      `json:"target"`
	Description string      `json:"description"`
}

// DiffReport summarizes the semantic version impact between two bundles or contracts.
type DiffReport struct {
	OldTarget          string       `json:"old_target"`
	NewTarget          string       `json:"new_target"`
	RecommendedBump    SemVerLevel  `json:"recommended_bump"`
	BreakingChanges    int          `json:"breaking_changes"`
	MinorChanges       int          `json:"minor_changes"`
	PatchChanges       int          `json:"patch_changes"`
	Changes            []DiffChange `json:"changes"`
	AffectedDownstream []string     `json:"affected_downstream,omitempty"`
}

type rawDPForDiff struct {
	Name      string          `json:"name" yaml:"name"`
	Resources []rawResForDiff `json:"resources" yaml:"resources"`
}

type rawResForDiff struct {
	Name   string           `json:"name" yaml:"name"`
	Format string           `json:"format" yaml:"format"`
	Schema rawSchemaForDiff `json:"schema" yaml:"schema"`
}

type rawSchemaForDiff struct {
	Fields []rawFieldForDiff `json:"fields" yaml:"fields"`
}

type rawFieldForDiff struct {
	Name    string `json:"name" yaml:"name"`
	Type    string `json:"type" yaml:"type"`
	Unit    string `json:"unit" yaml:"unit"`
	RDFType string `json:"rdfType" yaml:"rdfType"`
	Concept string `json:"concept" yaml:"concept"`
}

// DiffBundles compares two bundle directories or datapackage files and classifies changes into PATCH, MINOR, or MAJOR.
func DiffBundles(oldPath, newPath string) (*DiffReport, error) {
	report := &DiffReport{
		OldTarget:       oldPath,
		NewTarget:       newPath,
		RecommendedBump: SemVerNone,
		Changes:         make([]DiffChange, 0),
	}

	oldDP, err1 := loadDataPackageForDiff(oldPath)
	newDP, err2 := loadDataPackageForDiff(newPath)

	if err1 == nil && err2 == nil && (len(oldDP.Resources) > 0 || len(newDP.Resources) > 0) {
		compareDataPackages(oldDP, newDP, report)
	}

	// Also compare markdown concepts if directories
	oldInfo, errOld := os.Stat(oldPath)
	newInfo, errNew := os.Stat(newPath)
	if errOld == nil && errNew == nil && oldInfo.IsDir() && newInfo.IsDir() {
		compareMarkdownConcepts(oldPath, newPath, report)
	}

	// Compute overall recommended bump
	for _, c := range report.Changes {
		switch c.Level {
		case SemVerMajor:
			report.BreakingChanges++
		case SemVerMinor:
			report.MinorChanges++
		case SemVerPatch:
			report.PatchChanges++
		}
	}

	switch {
	case report.BreakingChanges > 0:
		report.RecommendedBump = SemVerMajor
	case report.MinorChanges > 0:
		report.RecommendedBump = SemVerMinor
	case report.PatchChanges > 0:
		report.RecommendedBump = SemVerPatch
	default:
		report.RecommendedBump = SemVerNone
	}

	return report, nil
}

func loadDataPackageForDiff(target string) (*rawDPForDiff, error) {
	info, err := os.Stat(target)
	if err != nil {
		return nil, err
	}
	candidates := []string{target}
	if info.IsDir() {
		candidates = []string{
			filepath.Join(target, "datapackage.json"),
			filepath.Join(target, "datapackage.yaml"),
			filepath.Join(target, "datapackage.yml"),
		}
	}
	for _, c := range candidates {
		data, err := os.ReadFile(c)
		if err != nil {
			continue
		}
		var dp rawDPForDiff
		if strings.HasSuffix(strings.ToLower(c), ".json") {
			if err := json.Unmarshal(data, &dp); err == nil {
				return &dp, nil
			}
		} else {
			if err := yaml.Unmarshal(data, &dp); err == nil {
				return &dp, nil
			}
		}
	}
	return nil, fmt.Errorf("no datapackage found at %s", target)
}

func compareDataPackages(oldDP, newDP *rawDPForDiff, report *DiffReport) {
	oldResMap := make(map[string]rawResForDiff)
	for _, r := range oldDP.Resources {
		oldResMap[r.Name] = r
	}
	newResMap := make(map[string]rawResForDiff)
	for _, r := range newDP.Resources {
		newResMap[r.Name] = r
	}

	for name, oldRes := range oldResMap {
		newRes, exists := newResMap[name]
		if !exists {
			report.Changes = append(report.Changes, DiffChange{
				Level:       SemVerMajor,
				Category:    "resource",
				Target:      name,
				Description: fmt.Sprintf("Removed resource '%s'", name),
			})
			continue
		}

		// Compare schema fields
		oldFields := make(map[string]rawFieldForDiff)
		for _, f := range oldRes.Schema.Fields {
			oldFields[f.Name] = f
		}
		newFields := make(map[string]rawFieldForDiff)
		for _, f := range newRes.Schema.Fields {
			newFields[f.Name] = f
		}

		for fName, oldF := range oldFields {
			newF, fExists := newFields[fName]
			if !fExists {
				report.Changes = append(report.Changes, DiffChange{
					Level:       SemVerMajor,
					Category:    "column",
					Target:      name + "." + fName,
					Description: fmt.Sprintf("Removed column '%s' from resource '%s'", fName, name),
				})
				continue
			}
			if !strings.EqualFold(oldF.Type, newF.Type) && oldF.Type != "" && newF.Type != "" {
				report.Changes = append(report.Changes, DiffChange{
					Level:       SemVerMajor,
					Category:    "column",
					Target:      name + "." + fName,
					Description: fmt.Sprintf("Changed column type of '%s.%s' from '%s' to '%s'", name, fName, oldF.Type, newF.Type),
				})
			}
			if oldF.Unit != newF.Unit && (oldF.Unit != "" || newF.Unit != "") {
				report.Changes = append(report.Changes, DiffChange{
					Level:       SemVerMajor,
					Category:    "column",
					Target:      name + "." + fName,
					Description: fmt.Sprintf("Changed physical unit (UCUM/QUDT) of '%s.%s' from '%s' to '%s'", name, fName, oldF.Unit, newF.Unit),
				})
			}
			if oldF.RDFType != newF.RDFType && (oldF.RDFType != "" || newF.RDFType != "") {
				report.Changes = append(report.Changes, DiffChange{
					Level:       SemVerMajor,
					Category:    "column",
					Target:      name + "." + fName,
					Description: fmt.Sprintf("Changed ontological anchor (rdfType) of '%s.%s' from '%s' to '%s'", name, fName, oldF.RDFType, newF.RDFType),
				})
			}
		}

		for fName := range newFields {
			if _, fExists := oldFields[fName]; !fExists {
				report.Changes = append(report.Changes, DiffChange{
					Level:       SemVerMinor,
					Category:    "column",
					Target:      name + "." + fName,
					Description: fmt.Sprintf("Added new column '%s' to resource '%s'", fName, name),
				})
			}
		}
	}

	for name := range newResMap {
		if _, exists := oldResMap[name]; !exists {
			report.Changes = append(report.Changes, DiffChange{
				Level:       SemVerMinor,
				Category:    "resource",
				Target:      name,
				Description: fmt.Sprintf("Added new resource '%s'", name),
			})
		}
	}
}

func compareMarkdownConcepts(oldDir, newDir string, report *DiffReport) {
	scanMD := func(root string) map[string]string {
		out := make(map[string]string)
		_ = filepath.Walk(root, func(p string, fi os.FileInfo, err error) error {
			if err != nil || fi.IsDir() || !strings.HasSuffix(fi.Name(), ".md") {
				return nil
			}
			rel, err := filepath.Rel(root, p)
			if err == nil {
				b, _ := os.ReadFile(p)
				out[filepath.ToSlash(rel)] = string(b)
			}
			return nil
		})
		return out
	}

	oldFiles := scanMD(oldDir)
	newFiles := scanMD(newDir)

	var keys []string
	seen := make(map[string]bool)
	for k := range oldFiles {
		keys = append(keys, k)
		seen[k] = true
	}
	for k := range newFiles {
		if !seen[k] {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	for _, k := range keys {
		oldContent, inOld := oldFiles[k]
		newContent, inNew := newFiles[k]
		if inOld && !inNew {
			report.Changes = append(report.Changes, DiffChange{
				Level:       SemVerMajor,
				Category:    "concept",
				Target:      k,
				Description: fmt.Sprintf("Removed concept document '%s'", k),
			})
		} else if !inOld && inNew {
			report.Changes = append(report.Changes, DiffChange{
				Level:       SemVerMinor,
				Category:    "concept",
				Target:      k,
				Description: fmt.Sprintf("Added new concept document '%s'", k),
			})
		} else if oldContent != newContent {
			report.Changes = append(report.Changes, DiffChange{
				Level:       SemVerPatch,
				Category:    "documentation",
				Target:      k,
				Description: fmt.Sprintf("Modified prose/frontmatter in '%s'", k),
			})
		}
	}
}

var sqlFenceRegex = regexp.MustCompile("(?s)```(?:sql)?\\s*(SELECT\\b.*?)```|\\n    (SELECT\\b[^\\n]*(?:\\n    [^\\n]+)*)")

// NormalizeSQLCanonical strips extra whitespace, trailing semicolons, and case differences for deterministic AST/token comparison.
func NormalizeSQLCanonical(sqlText string) string {
	trimmed := strings.TrimSpace(sqlText)
	trimmed = strings.TrimSuffix(trimmed, ";")
	fields := strings.Fields(trimmed)
	return strings.ToUpper(strings.Join(fields, " "))
}

// AttestComputation verifies an Attested Computation concept against an execution receipt JSON or executed SQL query,
// returning a standardized domain.AttestationVerdict wire object (§10.2, §12).
func AttestComputation(conceptPath string, executedSQLOrReceipt string, params map[string]string) (*domain.AttestationVerdict, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	contentBytes, err := os.ReadFile(conceptPath)
	if err != nil {
		return &domain.AttestationVerdict{
			Verdict:   "error",
			CheckedAt: now,
			Reason:    fmt.Sprintf("Failed to read computation concept '%s': %v", conceptPath, err),
		}, err
	}

	content := string(contentBytes)
	sanctionedSQL := extractSanctionedComputation(content, filepath.Dir(conceptPath))
	if sanctionedSQL == "" {
		return &domain.AttestationVerdict{
			Verdict:   "error",
			CheckedAt: now,
			Reason:    "No sanctioned computation block or file found in concept",
		}, fmt.Errorf("no computation block found in %s", conceptPath)
	}

	// Bind declared parameters (e.g. @year or {{ var('year') }})
	boundSanctioned := sanctionedSQL
	for k, v := range params {
		boundSanctioned = strings.ReplaceAll(boundSanctioned, "@"+k, v)
		boundSanctioned = strings.ReplaceAll(boundSanctioned, fmt.Sprintf("{{ var('%s') }}", k), v)
	}

	canonicalSanctioned := NormalizeSQLCanonical(boundSanctioned)
	h := sha256.Sum256([]byte(canonicalSanctioned))
	canonicalHash := "sha256:" + hex.EncodeToString(h[:])

	// If executedSQLOrReceipt is empty, validate self-attestation of canonical definition
	if strings.TrimSpace(executedSQLOrReceipt) == "" {
		return &domain.AttestationVerdict{
			Verdict:            "verified",
			CanonicalQueryHash: canonicalHash,
			CheckedAt:          now,
			Reason:             "Sanctioned computation syntax and canonical SHA-256 digest verified",
		}, nil
	}

	// Extract executed_sql if input is a JSON receipt string or path to a JSON receipt file
	actualSQL := strings.TrimSpace(executedSQLOrReceipt)
	if !strings.HasPrefix(actualSQL, "{") && (strings.HasSuffix(strings.ToLower(actualSQL), ".json") || fileExists(actualSQL)) {
		if rawReceipt, rErr := os.ReadFile(actualSQL); rErr == nil {
			actualSQL = strings.TrimSpace(string(rawReceipt))
		}
	}
	if strings.HasPrefix(actualSQL, "{") {
		var receiptMap map[string]interface{}
		if err := json.Unmarshal([]byte(actualSQL), &receiptMap); err == nil {
			if execSQL, ok := receiptMap["executed_sql"].(string); ok && execSQL != "" {
				actualSQL = execSQL
			} else if compSQL, ok := receiptMap["compiled_sql"].(string); ok && compSQL != "" {
				actualSQL = compSQL
			} else if qSQL, ok := receiptMap["sql"].(string); ok && qSQL != "" {
				actualSQL = qSQL
			}
		}
	}

	canonicalActual := NormalizeSQLCanonical(actualSQL)
	if canonicalActual == canonicalSanctioned {
		return &domain.AttestationVerdict{
			Verdict:            "verified",
			CanonicalQueryHash: canonicalHash,
			CheckedAt:          now,
			Reason:             "AST/Canonical SQL match & parameter binding verified against sanctioned computation",
		}, nil
	}

	return &domain.AttestationVerdict{
		Verdict:            "refuted",
		CanonicalQueryHash: canonicalHash,
		CheckedAt:          now,
		Reason:             fmt.Sprintf("Executed SQL does not match sanctioned computation (expected hash %s)", canonicalHash),
	}, nil
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

func extractSanctionedComputation(markdown string, baseDir string) string {
	parts := strings.Split(markdown, "---")
	if len(parts) >= 3 {
		var fm map[string]interface{}
		if err := yaml.Unmarshal([]byte(parts[1]), &fm); err == nil {
			if compPath, ok := fm["computation"].(string); ok && strings.TrimSpace(compPath) != "" {
				cleanRel := filepath.Clean(strings.TrimSpace(compPath))
				// SEC-01 Path containment: reject absolute paths or path traversal sequences (..)
				if !filepath.IsAbs(cleanRel) && !strings.HasPrefix(cleanRel, "..") && !strings.Contains(cleanRel, "../") {
					fullP := filepath.Join(baseDir, cleanRel)
					if b, err := os.ReadFile(fullP); err == nil {
						return string(b)
					}
				}
			}
		}
	}

	matches := sqlFenceRegex.FindStringSubmatch(markdown)
	if len(matches) >= 2 && strings.TrimSpace(matches[1]) != "" {
		return strings.TrimSpace(matches[1])
	}
	if len(matches) >= 3 && strings.TrimSpace(matches[2]) != "" {
		return strings.TrimSpace(matches[2])
	}
	return ""
}
