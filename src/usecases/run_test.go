package usecases

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteScript(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "okf-run-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 1. Create a bundle in tmpDir
	err = InitBundle(tmpDir, "Run Test Bundle", "Testing okf run execution")
	if err != nil {
		t.Fatalf("InitBundle failed: %v", err)
	}

	// 2. Create a test bash script
	scriptPath := filepath.Join(tmpDir, "test_pipeline.sh")
	scriptContent := "#!/bin/bash\necho 'Processing dataset records...'\necho 'Done.'\nexit 0\n"
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("failed creating script: %v", err)
	}

	// 3. Execute script
	res, err := ExecuteScript(RunParams{
		ScriptPath: scriptPath,
		Target:     "test_bundle:sample_dataset",
		BundleDir:  tmpDir,
	})
	if err != nil {
		t.Fatalf("ExecuteScript failed: %v", err)
	}

	if !res.Success {
		t.Errorf("expected success, got exit code %d (stderr: %s)", res.ExitCode, res.Stderr)
	}
	if len(res.ScriptHash) != 64 {
		t.Errorf("expected 64-char sha256 hash, got '%s'", res.ScriptHash)
	}
	if !strings.Contains(res.Stdout, "Processing dataset records") {
		t.Errorf("stdout does not contain expected output: %s", res.Stdout)
	}

	// 4. Verify log.md has recorded execution
	logBytes, err := os.ReadFile(filepath.Join(tmpDir, "log.md"))
	if err != nil {
		t.Fatalf("failed reading log.md: %v", err)
	}
	if !strings.Contains(string(logBytes), "test_pipeline.sh") {
		t.Errorf("log.md does not contain script execution entry")
	}
}
