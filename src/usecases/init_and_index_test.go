package usecases

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/datosbolivia/open-knowledge-format/src/domain"
)

func TestInitBundle_CreatesValidBundle(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "okf_init_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	targetDir := filepath.Join(tmpDir, "my_new_bundle")
	err = InitBundle(targetDir, "My New Bundle", "Testing OKF v0.2 bundle initialization")
	if err != nil {
		t.Fatalf("InitBundle failed: %v", err)
	}

	// Verify files exist
	expectedFiles := []string{
		filepath.Join(targetDir, "index.md"),
		filepath.Join(targetDir, "log.md"),
		filepath.Join(targetDir, "concepts", "sample_dataset.md"),
		filepath.Join(targetDir, "concepts", "datapackage.json"),
	}
	for _, ef := range expectedFiles {
		if _, err := os.Stat(ef); os.IsNotExist(err) {
			t.Errorf("expected file '%s' does not exist", ef)
		}
	}

	// Validate the newly initialized bundle using the domain validator
	report, err := domain.ValidateBundle(targetDir)
	if err != nil {
		t.Fatalf("ValidateBundle failed on initialized bundle: %v", err)
	}
	if !report.Valid {
		t.Errorf("initialized bundle must be valid out-of-the-box, but got %d errors", report.TotalErrors)
		for _, issue := range report.Issues {
			t.Logf("Issue: [%s] %s (%s)", issue.Severity, issue.Message, issue.Code)
		}
	}
}

func TestGenerateIndex_ScansAndWritesTable(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "okf_index_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create 2 test concepts
	c1 := `---
type: dataset
title: Air Quality 2024
description: PM2.5 and PM10 hourly records
status: active
---
# Air Quality
`
	c2 := `---
type: model
title: Dengue Forecast
description: Logistic regression model for dengue outbreaks
status: proposed
---
# Dengue Model
`
	subDir := filepath.Join(tmpDir, "sub")
	_ = os.MkdirAll(subDir, 0755)

	_ = os.WriteFile(filepath.Join(tmpDir, "air.md"), []byte(c1), 0644)
	_ = os.WriteFile(filepath.Join(subDir, "dengue.md"), []byte(c2), 0644)

	report, err := GenerateIndex(tmpDir)
	if err != nil {
		t.Fatalf("GenerateIndex failed: %v", err)
	}

	if report.ConceptsIndexed != 2 {
		t.Errorf("expected 2 concepts indexed, got %d", report.ConceptsIndexed)
	}

	// Check index.md content
	indexPath := filepath.Join(tmpDir, "index.md")
	contentBytes, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("failed to read generated index.md: %v", err)
	}
	content := string(contentBytes)

	if !strings.Contains(content, "Air Quality 2024") {
		t.Errorf("index.md missing 'Air Quality 2024'")
	}
	if !strings.Contains(content, "Dengue Forecast") {
		t.Errorf("index.md missing 'Dengue Forecast'")
	}
	if !strings.Contains(content, "type: index") {
		t.Errorf("index.md missing 'type: index' frontmatter")
	}
}
