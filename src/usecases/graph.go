package usecases

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/datosbolivia/open-knowledge-format/src/domain"
	"gopkg.in/yaml.v3"
)

type GraphEdge struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	Relation string `json:"relation"`
}

type GraphAdjacencyReport struct {
	BundleDir string           `json:"bundle_dir"`
	Nodes     []ConceptSummary `json:"nodes"`
	Edges     []GraphEdge      `json:"edges"`
}

var graphLinkWithLabelRegex = regexp.MustCompile(`\[([^\]]+)\]\(([^)#\s]+\.md)(?:#[^)]*)?\)`)

// BuildGraphAdjacency extracts nodes and directed edges between concepts in the bundle.
func BuildGraphAdjacency(bundleDir string) (*GraphAdjacencyReport, error) {
	cleanDir := filepath.Clean(bundleDir)
	info, err := os.Stat(cleanDir)
	if err != nil {
		return nil, fmt.Errorf("bundle directory error: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("target '%s' is not a directory", cleanDir)
	}

	report := &GraphAdjacencyReport{
		BundleDir: cleanDir,
		Nodes:     []ConceptSummary{},
		Edges:     []GraphEdge{},
	}

	nodeMap := make(map[string]ConceptSummary)

	err = filepath.Walk(cleanDir, func(path string, fi os.FileInfo, walkErr error) error {
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

		lower := strings.ToLower(baseName)
		if lower == "log.md" || lower == "readme.md" {
			return nil
		}

		relPath, err := filepath.Rel(cleanDir, path)
		if err != nil {
			relPath = path
		}

		contentBytes, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := string(contentBytes)

		var fm domain.ConceptFrontmatter
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
				_ = yaml.Unmarshal([]byte(strings.Join(lines[1:endIdx], "\n")), &fm)
			}
		}

		title := strings.TrimSpace(fm.Title)
		if title == "" {
			title = strings.TrimSuffix(baseName, ".md")
		}

		cType := strings.TrimSpace(fm.Type)
		if cType == "" {
			if lower == "index.md" {
				cType = "index"
			} else {
				cType = "concept"
			}
		}

		status := strings.TrimSpace(fm.Status)
		if status == "" {
			status = "active"
		}

		summary := ConceptSummary{
			RelativePath: relPath,
			Type:         cType,
			Title:        title,
			Description:  strings.TrimSpace(fm.Description),
			Status:       status,
			Tags:         fm.Tags,
		}

		nodeMap[relPath] = summary

		fileDir := filepath.Dir(path)
		matches := graphLinkWithLabelRegex.FindAllStringSubmatch(content, -1)
		for _, m := range matches {
			if len(m) < 3 {
				continue
			}
			label := strings.TrimSpace(m[1])
			rawTarget := m[2]
			resolved := filepath.Clean(filepath.Join(fileDir, rawTarget))
			targetRel, err := filepath.Rel(cleanDir, resolved)
			if err != nil {
				targetRel = resolved
			}

			report.Edges = append(report.Edges, GraphEdge{
				Source:   relPath,
				Target:   targetRel,
				Relation: label,
			})
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed walking graph: %w", err)
	}

	for _, node := range nodeMap {
		report.Nodes = append(report.Nodes, node)
	}

	sort.Slice(report.Nodes, func(i, j int) bool {
		return report.Nodes[i].RelativePath < report.Nodes[j].RelativePath
	})

	sort.Slice(report.Edges, func(i, j int) bool {
		if report.Edges[i].Source == report.Edges[j].Source {
			return report.Edges[i].Target < report.Edges[j].Target
		}
		return report.Edges[i].Source < report.Edges[j].Source
	})

	return report, nil
}
