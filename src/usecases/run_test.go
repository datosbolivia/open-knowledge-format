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

func TestExecuteScript_ProvenanceSyncDriftAndReport(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "okf-prov-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := InitBundle(tmpDir, "Provenance Bundle", "Testing safe distribution sync and drift"); err != nil {
		t.Fatalf("InitBundle failed: %v", err)
	}

	// Curated datapackage.yaml with human concept anchor + description
	curatedPkg := `name: prov_bundle
profile: tabular-data-package
resources:
  - name: ventas
    path: data/old_ventas.csv
    format: csv
    schema:
      fields:
        - name: municipio
          type: string
          description: "Codigo municipal curado"
          concept: "concepts/municipio.md"
        - name: monto
          type: number
        - name: legacy_col
          type: string
`
	if err := os.WriteFile(filepath.Join(tmpDir, "datapackage.yaml"), []byte(curatedPkg), 0644); err != nil {
		t.Fatalf("failed writing datapackage.yaml: %v", err)
	}

	reportPath := filepath.Join(tmpDir, "run_report.md")
	scriptPath := filepath.Join(tmpDir, "etl.sh")
	scriptBody := `#!/bin/bash
cat << 'EOF' >> "$OKF_PROVENANCE_OUT"
{"type":"input","uri":"raw:ventas_brutas","format":"csv"}
{"type":"step","step":"normalize_sales","tool":"bash","duration_ms":12,"status":"ok"}
{"type":"quality","resource":"ventas","row_count":150,"completeness":0.98}
{"type":"output","resource":"ventas","path":"data/ventas_2026.parquet","format":"parquet","hash_sha256":"abc123def456","bytes_count":2048,"row_count":150,"completeness":0.98,"fields":[{"name":"municipio","type":"string"},{"name":"monto","type":"string"},{"name":"nuevo_campo","type":"integer"}]}
{"type":"note","message":"Automated run completed cleanly."}
EOF
`
	if err := os.WriteFile(scriptPath, []byte(scriptBody), 0755); err != nil {
		t.Fatalf("failed writing etl.sh: %v", err)
	}

	res, err := ExecuteScript(RunParams{
		ScriptPath: scriptPath,
		Target:     "prov_bundle:ventas",
		BundleDir:  tmpDir,
		ReportPath: reportPath,
	})
	if err != nil {
		t.Fatalf("ExecuteScript failed: %v", err)
	}
	if !res.Success {
		t.Fatalf("expected success, stderr: %s", res.Stderr)
	}
	if len(res.SyncedResources) != 1 || res.SyncedResources[0] != "ventas" {
		t.Errorf("expected synced resource 'ventas', got %v", res.SyncedResources)
	}
	if len(res.DriftAlerts) != 2 {
		t.Fatalf("expected 2 drift alerts (type_changed + missing_column), got %d: %+v", len(res.DriftAlerts), res.DriftAlerts)
	}

	// Verify curated concept anchor & legacy_col were preserved, path/hash updated, and nuevo_campo appended
	updatedPkgBytes, err := os.ReadFile(filepath.Join(tmpDir, "datapackage.yaml"))
	if err != nil {
		t.Fatalf("failed reading updated datapackage.yaml: %v", err)
	}
	updatedPkg := string(updatedPkgBytes)
	for _, want := range []string{
		"data/ventas_2026.parquet",
		"abc123def456",
		"concepts/municipio.md",
		"Codigo municipal curado",
		"legacy_col",
		"nuevo_campo",
	} {
		if !strings.Contains(updatedPkg, want) {
			t.Errorf("updated datapackage.yaml missing %q:\n%s", want, updatedPkg)
		}
	}

	// Verify log.md records drift alerts
	logBytes, _ := os.ReadFile(filepath.Join(tmpDir, "log.md"))
	if !strings.Contains(string(logBytes), "DRIFT [WARNING]") {
		t.Errorf("expected DRIFT alerts in log.md, got:\n%s", string(logBytes))
	}

	// Verify optional Markdown run report
	repBytes, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("expected Markdown report at %s: %v", reportPath, err)
	}
	if !strings.Contains(string(repBytes), "## Drift Alerts") || !strings.Contains(string(repBytes), "normalize_sales") {
		t.Errorf("unexpected Markdown report content:\n%s", string(repBytes))
	}
}

func TestBuildNotebookDelegateCommand(t *testing.T) {
	cmd, args := buildNotebookDelegateCommand("etl.ipynb", RunParams{
		Target: "ds:res",
		NoSync: true,
		Params: map[string]string{"year": "2026"},
	})
	joined := cmd + " " + strings.Join(args, " ")
	if !strings.Contains(joined, "run etl.ipynb") || !strings.Contains(joined, "--target ds:res") ||
		!strings.Contains(joined, "--no-sync") || !strings.Contains(joined, "-p year=2026") {
		t.Errorf("unexpected notebook delegation command: %s", joined)
	}
}
