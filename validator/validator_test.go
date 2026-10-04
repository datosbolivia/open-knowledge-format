package validator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateConceptContentValid(t *testing.T) {
	content := `---
type: dataset
title: Test Dataset
description: A clean test dataset
skos:
  - https://www.wikidata.org/entity/Q39825
contracts:
  - type: datapackage
    path: https://example.org/datapackage.json
policy: allow_all
---

# Test Dataset Body
Narrative goes here.
`
	report := ValidateConceptContent(content, "test.md")
	if !report.Valid {
		t.Fatalf("Expected valid report, got %d errors: %+v", report.TotalErrors, report.Issues)
	}
	if report.TotalErrors != 0 {
		t.Errorf("Expected 0 errors, got %d", report.TotalErrors)
	}
}

func TestValidateConceptMissingType(t *testing.T) {
	content := `---
title: Missing Type Dataset
---
Body text.
`
	report := ValidateConceptContent(content, "invalid.md")
	if report.Valid {
		t.Fatal("Expected report to be invalid due to missing type")
	}
	hasTypeErr := false
	for _, issue := range report.Issues {
		if issue.Code == "PE2-TYPE-REQUIRED" {
			hasTypeErr = true
		}
	}
	if !hasTypeErr {
		t.Errorf("Expected PE2-TYPE-REQUIRED issue, got: %+v", report.Issues)
	}
}

func TestValidateTriad(t *testing.T) {
	validTriads := []string{
		"bolivia:elecciones:votos",
		"air_quality:compilacion_datos",
		"p2p-bob:exchange/trades",
		"'bolivia:censo:poblacion'",
		`"bolivia/elecciones/votos"`,
	}

	for _, tr := range validTriads {
		r := ValidateTriad(tr)
		if !r.Valid {
			t.Errorf("Expected triad '%s' to be valid, got errors: %+v", tr, r.Issues)
		}
	}

	invalidTriads := []string{
		"",
		"single_word",
		"invalid:::format",
	}

	for _, tr := range invalidTriads {
		r := ValidateTriad(tr)
		if r.Valid {
			t.Errorf("Expected triad '%s' to be invalid, but passed", tr)
		}
	}
}

func TestValidateBundle(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "okf-bundle-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create index.md
	_ = os.WriteFile(filepath.Join(tmpDir, "index.md"), []byte("# Index\nDirectory listing"), 0644)

	// Create a valid concept
	concept := `---
type: dataset
title: Sample
skos:
  - https://schema.org/Dataset
contracts:
  - type: datapackage
    path: https://example.com/dp.json
---
Body
`
	_ = os.WriteFile(filepath.Join(tmpDir, "concept1.md"), []byte(concept), 0644)

	report, err := ValidateBundle(tmpDir)
	if err != nil {
		t.Fatalf("ValidateBundle error: %v", err)
	}
	if !report.Valid {
		t.Errorf("Expected bundle to be valid, got errors: %+v", report.Issues)
	}
}
