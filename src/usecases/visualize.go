package usecases

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed assets/*
var vizAssets embed.FS

var (
	linkRegex = regexp.MustCompile(`\]\(([^)\s]+\.md)(?:#[A-Za-z0-9_\-]*)?\)`)
	typePalette = map[string]string{
		"BigQuery Dataset": "#8b5cf6",
		"BigQuery Table":   "#3b82f6",
		"Reference":        "#10b981",
		"dcat:Dataset":     "#3b82f6",
		"dcat:DataService": "#f59e0b",
		"dcat:Distribution": "#10b981",
		"dcat:CatalogRecord": "#8b5cf6",
	}
	defaultNodeColor = "#94a3b8"
)

type ConceptNode struct {
	ID          string                 `json:"id"`
	Label       string                 `json:"label"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Resource    string                 `json:"resource"`
	Tags        []string               `json:"tags"`
	Status      string                 `json:"status"`
	Generated   map[string]interface{} `json:"generated"`
	Verified    []interface{}          `json:"verified"`
	StaleAfter  string                 `json:"stale_after"`
	Sources     []interface{}          `json:"sources"`
	TrustTier   string                 `json:"trust_tier"`
	Stale       bool                   `json:"stale"`
	Color       string                 `json:"color"`
	Size        int                    `json:"size"`
}

type CytoscapeNode struct {
	Data ConceptNode `json:"data"`
}

type CytoscapeEdgeData struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

type CytoscapeEdge struct {
	Data CytoscapeEdgeData `json:"data"`
}

type GraphPayload struct {
	Nodes   []CytoscapeNode   `json:"nodes"`
	Edges   []CytoscapeEdge   `json:"edges"`
	Bodies  map[string]string `json:"bodies"`
	Types   []string          `json:"types"`
	Palette map[string]string `json:"palette"`
}

type VisualizeStats struct {
	ConceptsCount int
	EdgesCount    int
	BytesWritten  int
}

type VisualizeUseCase struct{}

func NewVisualizeUseCase() *VisualizeUseCase {
	return &VisualizeUseCase{}
}

func (uc *VisualizeUseCase) Generate(bundleRoot string, outPath string, bundleName string) (*VisualizeStats, error) {
	absRoot, err := filepath.Abs(bundleRoot)
	if err != nil {
		return nil, fmt.Errorf("invalid bundle root: %w", err)
	}

	info, err := os.Stat(absRoot)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("bundle directory not found: %s", absRoot)
	}

	if bundleName == "" {
		bundleName = filepath.Base(absRoot)
	}

	type conceptEntry struct {
		node    ConceptNode
		body    string
		linksTo []string
	}

	var entries []conceptEntry
	conceptIDs := make(map[string]bool)

	err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return walkErr
		}
		if !strings.HasSuffix(d.Name(), ".md") || d.Name() == "index.md" {
			return nil
		}

		relPath, err := filepath.Rel(absRoot, path)
		if err != nil {
			return nil
		}

		conceptID := strings.TrimSuffix(relPath, ".md")
		conceptID = filepath.ToSlash(conceptID)

		contentBytes, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		parts := strings.Split(string(contentBytes), "---")
		if len(parts) < 3 {
			return nil
		}

		fmText := parts[1]
		bodyText := strings.TrimSpace(strings.Join(parts[2:], "---"))

		var fm map[string]interface{}
		if err := yaml.Unmarshal([]byte(fmText), &fm); err != nil {
			return nil
		}
		if fm == nil {
			fm = make(map[string]interface{})
		}

		cType := getString(fm, "type", "Unknown")
		title := getString(fm, "title", conceptID)
		desc := getString(fm, "description", "")
		res := getString(fm, "resource", "")
		status := getString(fm, "status", "stable")
		staleAfter := getString(fm, "stale_after", "")

		var tags []string
		if rawTags, ok := fm["tags"].([]interface{}); ok {
			for _, t := range rawTags {
				tags = append(tags, fmt.Sprint(t))
			}
		}

		var genMap map[string]interface{}
		if gm, ok := fm["generated"].(map[string]interface{}); ok {
			genMap = gm
		} else {
			genMap = make(map[string]interface{})
		}

		var verifiedList []interface{}
		if vList, ok := fm["verified"].([]interface{}); ok {
			verifiedList = vList
		} else if vMap, ok := fm["verified"].(map[string]interface{}); ok {
			verifiedList = []interface{}{vMap}
		}

		var sourcesList []interface{}
		if sList, ok := fm["sources"].([]interface{}); ok {
			sourcesList = sList
		} else if sMap, ok := fm["sources"].(map[string]interface{}); ok {
			sourcesList = []interface{}{sMap}
		}

		// Trust tier
		tier := "unverified"
		if len(verifiedList) > 0 {
			hasHuman := false
			for _, v := range verifiedList {
				if vm, ok := v.(map[string]interface{}); ok {
					if by, ok := vm["by"].(string); ok && strings.HasPrefix(by, "human:") {
						hasHuman = true
						break
					}
				}
			}
			if hasHuman {
				tier = "human-reviewed"
			} else {
				tier = "machine-confirmed"
			}
		}

		// Stale check
		isStale := false
		if staleAfter != "" && strings.Contains(staleAfter, "T") {
			if t, err := time.Parse(time.RFC3339, staleAfter); err == nil {
				if time.Now().UTC().After(t) {
					isStale = true
				}
			}
		}

		color := typePalette[cType]
		if color == "" {
			color = defaultNodeColor
		}

		size := 30 + (len(bodyText) / 200)
		if size > 90 {
			size = 90
		}

		// Extract links in body
		var links []string
		docDir := filepath.Dir(path)
		for _, match := range linkRegex.FindAllStringSubmatch(bodyText, -1) {
			if len(match) > 1 {
				target := match[1]
				if strings.Contains(target, "://") || strings.HasPrefix(target, "/") {
					continue
				}
				targetAbs := filepath.Join(docDir, target)
				if relToRoot, err := filepath.Rel(absRoot, targetAbs); err == nil {
					cleanRel := strings.TrimSuffix(filepath.ToSlash(relToRoot), ".md")
					links = append(links, cleanRel)
				}
			}
		}

		node := ConceptNode{
			ID:          conceptID,
			Label:       title,
			Type:        cType,
			Description: desc,
			Resource:    res,
			Tags:        tags,
			Status:      status,
			Generated:   genMap,
			Verified:    verifiedList,
			StaleAfter:  staleAfter,
			Sources:     sourcesList,
			TrustTier:   tier,
			Stale:       isStale,
			Color:       color,
			Size:        size,
		}

		entries = append(entries, conceptEntry{node: node, body: bodyText, linksTo: links})
		conceptIDs[conceptID] = true
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed walking bundle: %w", err)
	}

	var nodes []CytoscapeNode
	var edges []CytoscapeEdge
	bodies := make(map[string]string)
	typeSet := make(map[string]bool)
	seenEdges := make(map[string]bool)

	for _, e := range entries {
		nodes = append(nodes, CytoscapeNode{Data: e.node})
		bodies[e.node.ID] = e.body
		typeSet[e.node.Type] = true

		for _, target := range e.linksTo {
			if target == e.node.ID || !conceptIDs[target] {
				continue
			}
			edgeKey := fmt.Sprintf("%s->%s", e.node.ID, target)
			if seenEdges[edgeKey] {
				continue
			}
			seenEdges[edgeKey] = true
			edges = append(edges, CytoscapeEdge{
				Data: CytoscapeEdgeData{
					ID:     fmt.Sprintf("%s__%s", e.node.ID, target),
					Source: e.node.ID,
					Target: target,
				},
			})
		}
	}

	var types []string
	for t := range typeSet {
		types = append(types, t)
	}
	sort.Strings(types)

	payload := GraphPayload{
		Nodes:   nodes,
		Edges:   edges,
		Bodies:  bodies,
		Types:   types,
		Palette: typePalette,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling graph payload: %w", err)
	}

	bundleNameJSON, _ := json.Marshal(bundleName)

	templateBytes, err := vizAssets.ReadFile("assets/viz.html")
	if err != nil {
		return nil, fmt.Errorf("failed reading embedded viz.html: %w", err)
	}

	cssBytes, err := vizAssets.ReadFile("assets/viz.css")
	if err != nil {
		return nil, fmt.Errorf("failed reading embedded viz.css: %w", err)
	}

	jsBytes, err := vizAssets.ReadFile("assets/viz.js")
	if err != nil {
		return nil, fmt.Errorf("failed reading embedded viz.js: %w", err)
	}

	html := string(templateBytes)
	html = strings.Replace(html, "/*__VIZ_CSS__*/", string(cssBytes), 1)
	html = strings.Replace(html, "/*__VIZ_JS__*/", string(jsBytes), 1)
	html = strings.Replace(html, "__BUNDLE_NAME__", string(bundleNameJSON), 1)
	html = strings.Replace(html, "__BUNDLE_DATA__", string(payloadJSON), 1)

	if outPath == "" {
		outPath = filepath.Join(absRoot, "viz.html")
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return nil, fmt.Errorf("failed creating output dir: %w", err)
	}

	if err := os.WriteFile(outPath, []byte(html), 0644); err != nil {
		return nil, fmt.Errorf("failed writing viz output: %w", err)
	}

	return &VisualizeStats{
		ConceptsCount: len(nodes),
		EdgesCount:    len(edges),
		BytesWritten:  len(html),
	}, nil
}

func getString(m map[string]interface{}, key string, def string) string {
	if val, ok := m[key]; ok && val != nil {
		return fmt.Sprint(val)
	}
	return def
}
