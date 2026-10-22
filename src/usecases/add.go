package usecases

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AddConceptParams parameters for creating a new concept node
type AddConceptParams struct {
	Path        string
	Type        string
	Title       string
	Description string
	Author      string
	Cadence     string
}

// AddConcept creates a new concept Markdown document with compliant frontmatter and progressive structure.
func AddConcept(params AddConceptParams) (string, error) {
	if params.Path == "" {
		return "", fmt.Errorf("concept path is required")
	}

	cleanPath := filepath.Clean(params.Path)
	if !strings.HasSuffix(cleanPath, ".md") {
		cleanPath += ".md"
	}

	if err := os.MkdirAll(filepath.Dir(cleanPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create concept directory: %w", err)
	}

	if _, err := os.Stat(cleanPath); err == nil {
		return "", fmt.Errorf("concept file already exists at %s", cleanPath)
	}

	cType := strings.TrimSpace(params.Type)
	if cType == "" {
		cType = "concept"
	}

	title := strings.TrimSpace(params.Title)
	if title == "" {
		base := filepath.Base(cleanPath)
		title = strings.TrimSuffix(base, ".md")
		title = strings.ReplaceAll(title, "_", " ")
		title = strings.ReplaceAll(title, "-", " ")
		title = strings.Title(title)
	}

	desc := strings.TrimSpace(params.Description)
	if desc == "" {
		desc = fmt.Sprintf("Autonomous knowledge node for %s.", title)
	}

	author := strings.TrimSpace(params.Author)
	if author == "" {
		author = "Agent <agent@datamesh.local>"
	}

	cadence := strings.TrimSpace(params.Cadence)
	if cadence == "" {
		cadence = "P1M"
	}

	now := time.Now().UTC()
	nowISO := now.Format(time.RFC3339)

	content := fmt.Sprintf(`---
type: %s
title: "%s"
description: "%s"
status: active
generated:
  by: "%s"
  at: "%s"
freshness:
  cadence: %s
---
# %s

%s

## Context & Invariants

## Operational Queries & Contracts

## Related Concepts
`, cType, title, desc, author, nowISO, cadence, title, desc)

	if err := os.WriteFile(cleanPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write concept file: %w", err)
	}

	bundleDir := findBundleRoot(filepath.Dir(cleanPath))
	if bundleDir != "" {
		_ = appendLogEntry(bundleDir, fmt.Sprintf("Added concept node '%s' (%s) via okf add.", title, filepath.Base(cleanPath)))
	}

	return cleanPath, nil
}
