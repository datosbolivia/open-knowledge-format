package usecases

import (
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

type AuditSeverity string

const (
	SeverityError   AuditSeverity = "ERROR"
	SeverityWarning AuditSeverity = "WARNING"
	SeverityInfo    AuditSeverity = "INFO"
)

type AuditIssue struct {
	File     string        `json:"file"`
	Severity AuditSeverity `json:"severity"`
	Rule     string        `json:"rule"`
	Message  string        `json:"message"`
}

type AuditReport struct {
	BundleDir    string       `json:"bundle_dir"`
	TotalFiles   int          `json:"total_files"`
	BrokenLinks  int          `json:"broken_links"`
	OrphanNodes  int          `json:"orphan_nodes"`
	CircularRefs int          `json:"circular_references"`
	StaleNodes   int          `json:"stale_nodes"`
	Issues       []AuditIssue `json:"issues"`
	Passed       bool         `json:"passed"`
}

var mdRelLinkRegex = regexp.MustCompile(`\[([^\]]*)\]\(([^)#\s]+\.md)(?:#[^)]*)?\)`)

// AuditBundle validates internal consistency, integrity of cross-references, freshness, and absence of broken links.
func AuditBundle(bundleDir string) (*AuditReport, error) {
	info, err := os.Stat(bundleDir)
	if err != nil {
		return nil, fmt.Errorf("bundle directory error: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("target '%s' is not a directory", bundleDir)
	}

	cleanBundleDir := filepath.Clean(bundleDir)

	type fileInfo struct {
		relPath   string
		fullPath  string
		content   string
		fm        *domain.ConceptFrontmatter
		hasFM     bool
		outLinks  map[string]bool
		inDegrees int
	}

	files := make(map[string]*fileInfo)

	err = filepath.Walk(cleanBundleDir, func(path string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if fi.IsDir() {
			if strings.HasPrefix(fi.Name(), ".") && fi.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}

		baseName := fi.Name()
		if !strings.HasSuffix(baseName, ".md") {
			return nil
		}

		relPath, err := filepath.Rel(cleanBundleDir, path)
		if err != nil {
			relPath = path
		}

		contentBytes, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := string(contentBytes)

		fiObj := &fileInfo{
			relPath:  relPath,
			fullPath: path,
			content:  content,
			outLinks: make(map[string]bool),
		}

		// Parse frontmatter if exists
		lines := strings.Split(content, "\n")
		if len(lines) >= 3 && strings.TrimSpace(lines[0]) == "---" {
			endIdx := -1
			for i := 1; i < len(lines); i++ {
				if strings.TrimSpace(lines[i]) == "---" {
					endIdx = i
					break
				}
			}
			if endIdx != -1 {
				var fm domain.ConceptFrontmatter
				if err := yaml.Unmarshal([]byte(strings.Join(lines[1:endIdx], "\n")), &fm); err == nil {
					fiObj.fm = &fm
					fiObj.hasFM = true
				}
			}
		}

		files[relPath] = fiObj
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed scanning bundle files: %w", err)
	}

	report := &AuditReport{
		BundleDir:  cleanBundleDir,
		TotalFiles: len(files),
		Passed:     true,
	}

	// 1. Analyze Links and Build Graph
	for relPath, f := range files {
		fileDir := filepath.Dir(f.fullPath)
		matches := mdRelLinkRegex.FindAllStringSubmatch(f.content, -1)
		for _, m := range matches {
			if len(m) < 3 {
				continue
			}
			rawTarget := m[2]
			// Resolve target relative to file's directory
			resolvedTarget := filepath.Clean(filepath.Join(fileDir, rawTarget))
			targetRel, err := filepath.Rel(cleanBundleDir, resolvedTarget)
			if err != nil {
				targetRel = resolvedTarget
			}

			// Check if target file exists
			if _, exists := files[targetRel]; !exists {
				// Target does not exist on disk
				if _, statErr := os.Stat(resolvedTarget); statErr != nil {
					report.BrokenLinks++
					report.Issues = append(report.Issues, AuditIssue{
						File:     relPath,
						Severity: SeverityError,
						Rule:     "BROKEN-LINK",
						Message:  fmt.Sprintf("Link points to non-existent target '%s'", rawTarget),
					})
					report.Passed = false
					continue
				}
			}

			f.outLinks[targetRel] = true
			if targetObj, ok := files[targetRel]; ok {
				targetObj.inDegrees++
			}
		}
	}

	// 2. Detect Circular References (RULE-12) & Self-loops
	checkedPairs := make(map[string]bool)
	for srcRel, f := range files {
		for dstRel := range f.outLinks {
			if srcRel == dstRel {
				report.CircularRefs++
				report.Issues = append(report.Issues, AuditIssue{
					File:     srcRel,
					Severity: SeverityWarning,
					Rule:     "RULE-12-SELF-REF",
					Message:  fmt.Sprintf("Concept has direct self-referential link to itself"),
				})
				continue
			}

			if targetObj, ok := files[dstRel]; ok {
				if targetObj.outLinks[srcRel] {
					pairKey := srcRel + "<->" + dstRel
					reverseKey := dstRel + "<->" + srcRel
					if !checkedPairs[pairKey] && !checkedPairs[reverseKey] {
						checkedPairs[pairKey] = true
						report.CircularRefs++
						report.Issues = append(report.Issues, AuditIssue{
							File:     srcRel,
							Severity: SeverityWarning,
							Rule:     "RULE-12-CIRCULAR-REF",
							Message:  fmt.Sprintf("Direct bidirectional circular link between '%s' and '%s'", srcRel, dstRel),
						})
					}
				}
			}
		}
	}

	// 3. Detect Orphan Nodes & Missing Frontmatter & Stale Nodes
	now := time.Now().UTC()
	for relPath, f := range files {
		baseName := filepath.Base(relPath)
		lower := strings.ToLower(baseName)
		isSystemFile := lower == "index.md" || lower == "log.md" || lower == "readme.md"

		// Frontmatter check
		if !isSystemFile && !f.hasFM {
			report.Issues = append(report.Issues, AuditIssue{
				File:     relPath,
				Severity: SeverityWarning,
				Rule:     "MISSING-FRONTMATTER",
				Message:  "Concept document missing valid YAML frontmatter with delimiters '---'",
			})
		}

		// Orphan node check
		if !isSystemFile {
			if len(f.outLinks) == 0 && f.inDegrees == 0 {
				report.OrphanNodes++
				report.Issues = append(report.Issues, AuditIssue{
					File:     relPath,
					Severity: SeverityWarning,
					Rule:     "ORPHAN-NODE",
					Message:  "Concept is completely disconnected (0 incoming links, 0 outgoing links)",
				})
			}
		}

		// Stale / Expired check
		if f.fm != nil && f.fm.Freshness != nil && f.fm.Freshness.Expires != "" {
			expiryStr := strings.TrimSpace(f.fm.Freshness.Expires)
			parsedTime, parseErr := time.Parse(time.RFC3339, expiryStr)
			if parseErr != nil {
				parsedTime, parseErr = time.Parse("2006-01-02", expiryStr)
			}
			if parseErr == nil && now.After(parsedTime) {
				report.StaleNodes++
				report.Issues = append(report.Issues, AuditIssue{
					File:     relPath,
					Severity: SeverityWarning,
					Rule:     "STALE-EXPIRED",
					Message:  fmt.Sprintf("Concept marked expired on %s", expiryStr),
				})
			}
		}
	}

	// Sort issues by file, then rule
	sort.Slice(report.Issues, func(i, j int) bool {
		if report.Issues[i].File == report.Issues[j].File {
			return report.Issues[i].Rule < report.Issues[j].Rule
		}
		return report.Issues[i].File < report.Issues[j].File
	})

	return report, nil
}
