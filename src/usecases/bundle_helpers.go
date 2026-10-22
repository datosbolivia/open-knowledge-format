package usecases

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// findBundleRoot traverses up from dir looking for index.md
func findBundleRoot(dir string) string {
	curr := filepath.Clean(dir)
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(curr, "index.md")); err == nil {
			return curr
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}
	return ""
}

// appendLogEntry appends a formatted entry to log.md under today's date heading
func appendLogEntry(bundleDir string, message string) error {
	logPath := filepath.Join(bundleDir, "log.md")
	now := time.Now().UTC()
	dateStr := now.Format("2006-01-02")
	heading := fmt.Sprintf("## %s", dateStr)
	entry := fmt.Sprintf("- %s\n", message)

	existingBytes, err := os.ReadFile(logPath)
	if err != nil {
		content := fmt.Sprintf("# Bundle Changelog\n\n%s\n%s", heading, entry)
		return os.WriteFile(logPath, []byte(content), 0644)
	}

	content := string(existingBytes)
	headingIdx := strings.Index(content, heading)
	if headingIdx != -1 {
		lineEnd := strings.Index(content[headingIdx:], "\n")
		if lineEnd != -1 {
			insertPos := headingIdx + lineEnd + 1
			content = content[:insertPos] + entry + content[insertPos:]
		} else {
			content += "\n" + entry
		}
	} else {
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += fmt.Sprintf("\n%s\n%s", heading, entry)
	}

	return os.WriteFile(logPath, []byte(content), 0644)
}
