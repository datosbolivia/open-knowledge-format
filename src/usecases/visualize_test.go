package usecases

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVisualizeUseCase(t *testing.T) {
	tmpDir := t.TempDir()

	// Create sample concept markdown
	c1Path := filepath.Join(tmpDir, "concept1.md")
	c1Content := `---
type: dcat:Dataset
title: Sample Dataset 1
description: First test dataset
tags: ["test", "demo"]
---
# Concept 1
Check out [Concept 2](concept2.md) for details.
`
	if err := os.WriteFile(c1Path, []byte(c1Content), 0644); err != nil {
		t.Fatalf("failed writing c1: %v", err)
	}

	c2Path := filepath.Join(tmpDir, "concept2.md")
	c2Content := `---
type: dcat:Distribution
title: Sample Distribution 2
description: Second test concept
---
# Concept 2
No outbound links.
`
	if err := os.WriteFile(c2Path, []byte(c2Content), 0644); err != nil {
		t.Fatalf("failed writing c2: %v", err)
	}

	uc := NewVisualizeUseCase()
	outPath := filepath.Join(tmpDir, "viz.html")
	stats, err := uc.Generate(tmpDir, outPath, "Test Bundle")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if stats.ConceptsCount != 2 {
		t.Errorf("expected 2 concepts, got %d", stats.ConceptsCount)
	}
	if stats.EdgesCount != 1 {
		t.Errorf("expected 1 edge, got %d", stats.EdgesCount)
	}

	outBytes, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed reading output: %v", err)
	}

	outStr := string(outBytes)
	if !strings.Contains(outStr, "Test Bundle") {
		t.Errorf("output does not contain bundle name")
	}
	if !strings.Contains(outStr, "Sample Dataset 1") {
		t.Errorf("output does not contain concept 1")
	}
	if !strings.Contains(outStr, "Sample Distribution 2") {
		t.Errorf("output does not contain concept 2")
	}
}
