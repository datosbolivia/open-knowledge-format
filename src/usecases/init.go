package usecases

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// InitBundle creates a standard, fully conformant OKF v0.2 bundle structure.
func InitBundle(targetDir string, bundleTitle string, description string) error {
	if bundleTitle == "" {
		bundleTitle = filepath.Base(targetDir)
	}
	if description == "" {
		description = fmt.Sprintf("Open Knowledge Format bundle for %s", bundleTitle)
	}

	conceptsDir := filepath.Join(targetDir, "concepts")
	if err := os.MkdirAll(conceptsDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	now := time.Now().UTC()
	nowISO := now.Format(time.RFC3339)
	dateStr := now.Format("2006-01-02")

	// 1. Root index.md (§8 Progressive Disclosure)
	indexContent := fmt.Sprintf(`---
type: index
title: "%s"
description: "%s"
status: active
generated:
  by: "okf CLI"
  at: "%s"
---
# %s

%s

## Concepts Directory
- [Sample Dataset](concepts/sample_dataset.md): Demonstrates OKF/ODKF v0.2 compliant metadata, provenance, and contracts.
`, bundleTitle, description, nowISO, bundleTitle, description)

	indexPath := filepath.Join(targetDir, "index.md")
	if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
		return fmt.Errorf("failed to write index.md: %w", err)
	}

	// 2. Append-only log.md (§9 Bundle Changelog)
	logContent := fmt.Sprintf(`# Bundle Changelog

## %s
- Initialized Open Knowledge Format bundle '%s' via okf CLI.
`, dateStr, bundleTitle)

	logPath := filepath.Join(targetDir, "log.md")
	if err := os.WriteFile(logPath, []byte(logContent), 0644); err != nil {
		return fmt.Errorf("failed to write log.md: %w", err)
	}

	// 3. Sample concept document
	conceptContent := fmt.Sprintf(`---
type: dataset
title: Sample Dataset
description: Demonstrates OKF/ODKF v0.2 compliant metadata, provenance, and contracts.
status: active
skos:
  - "http://www.w3.org/ns/dcat#Dataset"
contracts:
  - type: datapackage
    path: "datapackage.json"
generated:
  by: "Author Name <author@example.com>"
  at: "%s"
freshness:
  cadence: P1M
---
# Sample Dataset

Welcome to this Open Knowledge Format v0.2 concept document.
`, nowISO)

	conceptPath := filepath.Join(conceptsDir, "sample_dataset.md")
	if err := os.WriteFile(conceptPath, []byte(conceptContent), 0644); err != nil {
		return fmt.Errorf("failed to write sample_dataset.md: %w", err)
	}

	// 4. Companion local contract to guarantee out-of-the-box RULE-04 validity
	datapackageJSON := `{
  "name": "sample-dataset",
  "title": "Sample Dataset Contract",
  "version": "1.0.0",
  "resources": [
    {
      "name": "data",
      "path": "data.csv",
      "format": "csv"
    }
  ]
}
`
	dpPath := filepath.Join(conceptsDir, "datapackage.json")
	if err := os.WriteFile(dpPath, []byte(datapackageJSON), 0644); err != nil {
		return fmt.Errorf("failed to write datapackage.json: %w", err)
	}

	return nil
}
