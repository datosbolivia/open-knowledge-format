package usecases

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddAndLinkAndAudit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "okf-curate-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 1. Initialize bundle
	err = InitBundle(tmpDir, "Test Curation Bundle", "Curated test bundle")
	if err != nil {
		t.Fatalf("InitBundle failed: %v", err)
	}

	// 2. Add two new concepts
	c1Path := filepath.Join(tmpDir, "concepts", "water_quality.md")
	created1, err := AddConcept(AddConceptParams{
		Path:        c1Path,
		Type:        "dataset",
		Title:       "Water Quality Monitor",
		Description: "Real-time municipal water indicators",
		Author:      "Agent Hydro",
		Cadence:     "P1D",
	})
	if err != nil {
		t.Fatalf("AddConcept 1 failed: %v", err)
	}
	if created1 != c1Path {
		t.Errorf("expected %s, got %s", c1Path, created1)
	}

	c2Path := filepath.Join(tmpDir, "concepts", "health_indices.md")
	created2, err := AddConcept(AddConceptParams{
		Path:        c2Path,
		Type:        "concept",
		Title:       "Public Health Indices",
		Description: "Aggregate epidemiological indicators",
		Author:      "Agent Medic",
		Cadence:     "P1M",
	})
	if err != nil {
		t.Fatalf("AddConcept 2 failed: %v", err)
	}
	if created2 != c2Path {
		t.Errorf("expected %s, got %s", c2Path, created2)
	}

	// 3. Link water_quality -> health_indices
	err = LinkConcepts(c1Path, c2Path, "Correlates With")
	if err != nil {
		t.Fatalf("LinkConcepts failed: %v", err)
	}

	c1Content, err := os.ReadFile(c1Path)
	if err != nil {
		t.Fatalf("failed reading c1: %v", err)
	}
	if !strings.Contains(string(c1Content), "Correlates With") {
		t.Errorf("c1 does not contain link relation: %s", string(c1Content))
	}
	if !strings.Contains(string(c1Content), "health_indices.md") {
		t.Errorf("c1 does not contain relative link target: %s", string(c1Content))
	}

	// 4. Verify log.md has recorded operations
	logBytes, err := os.ReadFile(filepath.Join(tmpDir, "log.md"))
	if err != nil {
		t.Fatalf("failed reading log.md: %v", err)
	}
	logStr := string(logBytes)
	if !strings.Contains(logStr, "Water Quality Monitor") {
		t.Errorf("log.md missing water quality add entry")
	}
	if !strings.Contains(logStr, "Correlates With") {
		t.Errorf("log.md missing link entry")
	}

	// 5. Test Audit on clean bundle (except initial sample concept might be orphan)
	report, err := AuditBundle(tmpDir)
	if err != nil {
		t.Fatalf("AuditBundle failed: %v", err)
	}
	if report.TotalFiles < 3 {
		t.Errorf("expected >= 3 files, got %d", report.TotalFiles)
	}
	if report.BrokenLinks != 0 {
		t.Errorf("expected 0 broken links, got %d", report.BrokenLinks)
	}

	// 6. Test Audit detecting broken link
	brokenLinkContent := "---\ntype: concept\ntitle: Bad\nstatus: active\n---\n# Bad\n\n[Broken](non_existent_file.md)\n"
	badFilePath := filepath.Join(tmpDir, "concepts", "bad_link.md")
	if err := os.WriteFile(badFilePath, []byte(brokenLinkContent), 0644); err != nil {
		t.Fatalf("failed writing bad file: %v", err)
	}

	report2, err := AuditBundle(tmpDir)
	if err != nil {
		t.Fatalf("AuditBundle 2 failed: %v", err)
	}
	if report2.BrokenLinks != 1 {
		t.Errorf("expected 1 broken link, got %d", report2.BrokenLinks)
	}
	if report2.Passed {
		t.Errorf("expected audit to fail on broken link")
	}

	// 7. Test Graph Adjacency
	graphReport, err := BuildGraphAdjacency(tmpDir)
	if err != nil {
		t.Fatalf("BuildGraphAdjacency failed: %v", err)
	}
	if len(graphReport.Nodes) == 0 {
		t.Errorf("expected nodes in graph")
	}
	foundEdge := false
	for _, edge := range graphReport.Edges {
		if strings.Contains(edge.Source, "water_quality") && strings.Contains(edge.Target, "health_indices") {
			foundEdge = true
			if edge.Relation != "Correlates With" {
				t.Errorf("expected relation 'Correlates With', got '%s'", edge.Relation)
			}
		}
	}
	if !foundEdge {
		t.Errorf("expected to find edge water_quality -> health_indices in graph report")
	}
}
