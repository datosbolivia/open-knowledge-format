package usecases

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/datosbolivia/open-knowledge-format/src/domain"
)

type RunParams struct {
	ScriptPath string
	Target     string
	Args       []string
	Env        map[string]string
	BundleDir  string
	ReportPath string
	NoSync     bool
	Params     map[string]string
}

type RunExecutionResult struct {
	ID              string                   `json:"id"`
	ScriptPath      string                   `json:"script_path"`
	ScriptHash      string                   `json:"script_hash"`
	Target          string                   `json:"target"`
	ExitCode        int                      `json:"exit_code"`
	StartedAt       time.Time                `json:"started_at"`
	CompletedAt     time.Time                `json:"completed_at"`
	DurationMs      int64                    `json:"duration_ms"`
	Stdout          string                   `json:"stdout"`
	Stderr          string                   `json:"stderr"`
	Success         bool                     `json:"success"`
	Provenance      *domain.ProvenanceReport `json:"provenance,omitempty"`
	SyncedResources []string                 `json:"synced_resources,omitempty"`
	DriftAlerts     []domain.DriftAlert      `json:"drift_alerts,omitempty"`
	ReportPath      string                   `json:"report_path,omitempty"`
}

// ComputeFileSHA256 returns hex encoded SHA-256 checksum of a file.
func ComputeFileSHA256(path string) (string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:]), nil
}

// ExecuteScript executes a script (Python, Bash, or executable), capturing runtime metrics and code hash.
func ExecuteScript(params RunParams) (*RunExecutionResult, error) {
	cleanScript := filepath.Clean(params.ScriptPath)
	info, err := os.Stat(cleanScript)
	if err != nil {
		return nil, fmt.Errorf("script not found at '%s': %w", cleanScript, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("target '%s' is a directory, expected executable or script file", cleanScript)
	}

	codeHash, err := ComputeFileSHA256(cleanScript)
	if err != nil {
		return nil, fmt.Errorf("failed computing script hash: %w", err)
	}

	bDir := params.BundleDir
	if bDir == "" {
		bDir = findBundleRoot(filepath.Dir(cleanScript))
	}

	provFile, err := os.CreateTemp("", "okf-prov-*.jsonl")
	var provPath string
	if err == nil {
		provPath = provFile.Name()
		_ = provFile.Close()
		defer os.Remove(provPath)
	}

	ext := strings.ToLower(filepath.Ext(cleanScript))
	var cmdName string
	var cmdArgs []string

	switch ext {
	case ".py":
		cmdName = "python3"
		cmdArgs = append([]string{cleanScript}, params.Args...)
	case ".sh":
		cmdName = "bash"
		cmdArgs = append([]string{cleanScript}, params.Args...)
	case ".ipynb":
		cmdName, cmdArgs = buildNotebookDelegateCommand(cleanScript, params)
	default:
		cmdName = cleanScript
		cmdArgs = params.Args
	}

	execID := fmt.Sprintf("exec-%d", time.Now().UnixNano())
	startedAt := time.Now().UTC()

	cmd := exec.Command(cmdName, cmdArgs...)
	cmdEnv := os.Environ()
	if provPath != "" {
		cmdEnv = append(cmdEnv, "OKF_PROVENANCE_OUT="+provPath)
	}
	if bDir != "" {
		cmdEnv = append(cmdEnv, "OKF_BUNDLE_DIR="+bDir)
	}
	for k, v := range params.Params {
		cleanKey := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(k), "-", "_"))
		if cleanKey != "" {
			cmdEnv = append(cmdEnv, fmt.Sprintf("OKF_PARAM_%s=%s", cleanKey, v))
		}
	}
	for k, v := range params.Env {
		cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = cmdEnv

	var stdoutBuf, stderrBuf strings.Builder
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	runErr := cmd.Run()
	completedAt := time.Now().UTC()
	durationMs := completedAt.Sub(startedAt).Milliseconds()

	exitCode := 0
	if runErr != nil {
		if exitError, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	res := &RunExecutionResult{
		ID:          execID,
		ScriptPath:  cleanScript,
		ScriptHash:  codeHash,
		Target:      params.Target,
		ExitCode:    exitCode,
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		DurationMs:  durationMs,
		Stdout:      stdoutBuf.String(),
		Stderr:      stderrBuf.String(),
		Success:     exitCode == 0,
	}

	if provPath != "" {
		if provReport, pErr := parseProvenanceJSONL(provPath); pErr == nil && provReport != nil {
			res.Provenance = provReport
			res.DriftAlerts = append(res.DriftAlerts, provReport.Drifts...)
		}
	}

	// Safe additive distribution sync (updates path, hash, size, quality; appends new fields; never deletes curated knowledge)
	if res.Success && !params.NoSync && bDir != "" && res.Provenance != nil && len(res.Provenance.Outputs) > 0 {
		drifts, synced, syncErr := SyncDataPackageDistributions(bDir, res.Provenance.Outputs)
		if syncErr == nil {
			res.SyncedResources = synced
			res.DriftAlerts = append(res.DriftAlerts, drifts...)
		}
	}

	// Record concise provenance & drift summary in bundle log.md
	if bDir != "" {
		statusStr := "SUCCESS"
		if !res.Success {
			statusStr = fmt.Sprintf("FAILED (code %d)", exitCode)
		}
		targetStr := ""
		if params.Target != "" {
			targetStr = fmt.Sprintf(" for target '%s'", params.Target)
		}
		var extras []string
		if res.Provenance != nil {
			if len(res.Provenance.Inputs) > 0 {
				extras = append(extras, fmt.Sprintf("inputs: %d", len(res.Provenance.Inputs)))
			}
			if len(res.Provenance.Steps) > 0 {
				extras = append(extras, fmt.Sprintf("steps: %d", len(res.Provenance.Steps)))
			}
			if len(res.Provenance.Outputs) > 0 {
				extras = append(extras, fmt.Sprintf("outputs: %d", len(res.Provenance.Outputs)))
			}
		}
		if len(res.SyncedResources) > 0 {
			extras = append(extras, fmt.Sprintf("synced: [%s]", strings.Join(res.SyncedResources, ", ")))
		}
		if len(res.DriftAlerts) > 0 {
			extras = append(extras, fmt.Sprintf("DRIFT ALERTS: %d", len(res.DriftAlerts)))
		}
		extraStr := ""
		if len(extras) > 0 {
			extraStr = "; " + strings.Join(extras, ", ")
		}
		logMsg := fmt.Sprintf("Executed script '%s'%s [%s, duration: %dms, sha256: %s%s].",
			filepath.Base(cleanScript), targetStr, statusStr, durationMs, codeHash[:12], extraStr)
		_ = appendLogEntry(bDir, logMsg)
		for _, d := range res.DriftAlerts {
			_ = appendLogEntry(bDir, fmt.Sprintf("DRIFT [%s] %s: %s", d.Severity, d.Kind, d.Message))
		}
	}

	if strings.TrimSpace(params.ReportPath) != "" {
		reportMD := RenderRunReportMarkdown(res)
		if rErr := os.WriteFile(params.ReportPath, []byte(reportMD), 0644); rErr == nil {
			res.ReportPath = params.ReportPath
		}
	}

	return res, nil
}

func buildNotebookDelegateCommand(notebookPath string, params RunParams) (string, []string) {
	runnerCmd := "datamesh"
	var runnerArgs []string
	if _, err := exec.LookPath("datamesh"); err == nil {
		runnerArgs = []string{"run", notebookPath}
	} else {
		runnerCmd = "python3"
		runnerArgs = []string{"-m", "datamesh.cli", "run", notebookPath}
	}
	if params.Target != "" {
		runnerArgs = append(runnerArgs, "--target", params.Target)
	}
	if params.NoSync {
		runnerArgs = append(runnerArgs, "--no-sync")
	}
	if len(params.Params) > 0 {
		keys := make([]string, 0, len(params.Params))
		for k := range params.Params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			runnerArgs = append(runnerArgs, "-p", fmt.Sprintf("%s=%s", k, params.Params[k]))
		}
	}
	runnerArgs = append(runnerArgs, params.Args...)
	return runnerCmd, runnerArgs
}

func parseProvenanceJSONL(path string) (*domain.ProvenanceReport, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	report := &domain.ProvenanceReport{}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var ev domain.ProvenanceEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(ev.Type)) {
		case "input":
			report.Inputs = append(report.Inputs, ev)
		case "step":
			report.Steps = append(report.Steps, ev)
		case "output", "distribution":
			report.Outputs = append(report.Outputs, ev)
		case "quality":
			report.Quality = append(report.Quality, ev)
		case "metric":
			report.Metrics = append(report.Metrics, ev)
		case "note":
			report.Notes = append(report.Notes, ev)
		case "drift":
			sev := ev.Severity
			if sev == "" {
				sev = "WARNING"
			}
			report.Drifts = append(report.Drifts, domain.DriftAlert{
				Resource: ev.Resource,
				Kind:     "custom",
				Severity: sev,
				Message:  ev.Message,
			})
		}
	}
	return report, scanner.Err()
}

// RenderRunReportMarkdown builds an optional human-readable Markdown summary of an ETL execution.
func RenderRunReportMarkdown(res *RunExecutionResult) string {
	var sb strings.Builder
	status := "SUCCESS"
	if !res.Success {
		status = fmt.Sprintf("FAILED (exit %d)", res.ExitCode)
	}
	sb.WriteString(fmt.Sprintf("---\ntype: run_report\ntitle: \"ETL Run Report: %s\"\nstatus: %s\ntimestamp: \"%s\"\n---\n\n",
		filepath.Base(res.ScriptPath), strings.ToLower(strings.Split(status, " ")[0]), res.StartedAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("# ETL Execution Report: `%s`\n\n", filepath.Base(res.ScriptPath)))
	sb.WriteString(fmt.Sprintf("- **Execution ID**: `%s`\n", res.ID))
	sb.WriteString(fmt.Sprintf("- **Status**: `%s` (`%d ms`)\n", status, res.DurationMs))
	sb.WriteString(fmt.Sprintf("- **Script SHA-256**: `%s`\n", res.ScriptHash))
	if res.Target != "" {
		sb.WriteString(fmt.Sprintf("- **Target**: `%s`\n", res.Target))
	}
	if len(res.SyncedResources) > 0 {
		sb.WriteString(fmt.Sprintf("- **Synced Distributions (`datapackage.yaml`)**: `%s`\n", strings.Join(res.SyncedResources, "`, `")))
	}

	if res.Provenance != nil {
		if len(res.Provenance.Inputs) > 0 {
			sb.WriteString("\n## Inputs\n\n| Resource / URI | Format | Hash |\n| :--- | :--- | :--- |\n")
			for _, in := range res.Provenance.Inputs {
				uri := in.URI
				if uri == "" {
					uri = in.Path
				}
				sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` |\n", uri, in.Format, in.HashSHA256))
			}
		}
		if len(res.Provenance.Steps) > 0 {
			sb.WriteString("\n## Pipeline Steps\n\n| Step | Tool | Status | Duration (ms) |\n| :--- | :--- | :--- | :---: |\n")
			for _, st := range res.Provenance.Steps {
				stStatus := st.Status
				if stStatus == "" {
					stStatus = "ok"
				}
				sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | %d |\n", st.Step, st.Tool, stStatus, st.DurationMs))
			}
		}
		if len(res.Provenance.Quality) > 0 || len(res.Provenance.Metrics) > 0 {
			sb.WriteString("\n## Quality & Metrics\n\n")
			for _, q := range res.Provenance.Quality {
				sb.WriteString(fmt.Sprintf("- **Quality (`%s`)**: rows=%d, completeness=%.4f %s\n", q.Resource, q.RowCount, q.Completeness, q.Message))
			}
			for _, m := range res.Provenance.Metrics {
				sb.WriteString(fmt.Sprintf("- **Metric**: %s `%v`\n", m.Message, m.Metrics))
			}
		}
		if len(res.Provenance.Outputs) > 0 {
			sb.WriteString("\n## Output Distributions\n\n| Resource | Path | Format | Rows | Size (bytes) | SHA-256 |\n| :--- | :--- | :--- | :---: | :---: | :--- |\n")
			for _, out := range res.Provenance.Outputs {
				shortHash := out.HashSHA256
				if len(shortHash) > 12 {
					shortHash = shortHash[:12]
				}
				sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | %d | %d | `%s` |\n",
					out.Resource, out.Path, out.Format, out.RowCount, out.BytesCount, shortHash))
			}
		}
		if len(res.Provenance.Notes) > 0 {
			sb.WriteString("\n## Notes\n\n")
			for _, n := range res.Provenance.Notes {
				sb.WriteString(fmt.Sprintf("%s\n\n", n.Message))
			}
		}
	}

	if len(res.DriftAlerts) > 0 {
		sb.WriteString("\n## Drift Alerts\n\n")
		for _, d := range res.DriftAlerts {
			sb.WriteString(fmt.Sprintf("- **[%s] `%s`** (`%s`): %s\n", d.Severity, d.Kind, d.Resource, d.Message))
		}
	}

	return sb.String()
}
