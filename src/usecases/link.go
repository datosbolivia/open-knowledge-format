package usecases

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LinkConcepts links source Markdown document to target Markdown document.
func LinkConcepts(sourcePath, targetPath, relation string) error {
	sourceClean := filepath.Clean(sourcePath)
	targetClean := filepath.Clean(targetPath)

	if _, err := os.Stat(sourceClean); err != nil {
		return fmt.Errorf("source file does not exist: %w", err)
	}
	if _, err := os.Stat(targetClean); err != nil {
		return fmt.Errorf("target file does not exist: %w", err)
	}

	sourceDir := filepath.Dir(sourceClean)
	relPath, err := filepath.Rel(sourceDir, targetClean)
	if err != nil {
		return fmt.Errorf("failed to compute relative path: %w", err)
	}

	relLabel := strings.TrimSpace(relation)
	if relLabel == "" {
		targetBase := filepath.Base(targetClean)
		relLabel = strings.TrimSuffix(targetBase, ".md")
		relLabel = strings.ReplaceAll(relLabel, "_", " ")
		relLabel = strings.ReplaceAll(relLabel, "-", " ")
		relLabel = strings.Title(relLabel)
	}

	sourceBytes, err := os.ReadFile(sourceClean)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}
	content := string(sourceBytes)

	linkPattern := fmt.Sprintf("(%s)", relPath)
	if strings.Contains(content, linkPattern) {
		// Link already present, no-op
		return nil
	}

	newLink := fmt.Sprintf("- [%s](%s)\n", relLabel, relPath)

	relatedHeaderIdx := strings.Index(content, "## Related Concepts")
	if relatedHeaderIdx != -1 {
		headerEnd := strings.Index(content[relatedHeaderIdx:], "\n")
		if headerEnd != -1 {
			insertPos := relatedHeaderIdx + headerEnd + 1
			content = content[:insertPos] + newLink + content[insertPos:]
		} else {
			content += "\n" + newLink
		}
	} else {
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\n## Related Concepts\n" + newLink
	}

	if err := os.WriteFile(sourceClean, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write updated source file: %w", err)
	}

	bundleDir := findBundleRoot(filepath.Dir(sourceClean))
	if bundleDir != "" {
		_ = appendLogEntry(bundleDir, fmt.Sprintf("Linked '%s' -> '%s' (predicate: %s) via okf link.", filepath.Base(sourceClean), filepath.Base(targetClean), relLabel))
	}

	return nil
}
