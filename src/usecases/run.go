package usecases

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type RunParams struct {
	ScriptPath string
	Target     string
	Args       []string
	Env        map[string]string
	BundleDir  string
}

type RunExecutionResult struct {
	ID          string    `json:"id"`
	ScriptPath  string    `json:"script_path"`
	ScriptHash  string    `json:"script_hash"`
	Target      string    `json:"target"`
	ExitCode    int       `json:"exit_code"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
	DurationMs  int64     `json:"duration_ms"`
	Stdout      string    `json:"stdout"`
	Stderr      string    `json:"stderr"`
	Success     bool      `json:"success"`
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
	default:
		cmdName = cleanScript
		cmdArgs = params.Args
	}

	execID := fmt.Sprintf("exec-%d", time.Now().UnixNano())
	startedAt := time.Now().UTC()

	cmd := exec.Command(cmdName, cmdArgs...)
	if len(params.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range params.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

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

	// Record in bundle log.md if bundleDir provided
	bDir := params.BundleDir
	if bDir == "" {
		bDir = findBundleRoot(filepath.Dir(cleanScript))
	}
	if bDir != "" {
		statusStr := "SUCCESS"
		if !res.Success {
			statusStr = fmt.Sprintf("FAILED (code %d)", exitCode)
		}
		targetStr := ""
		if params.Target != "" {
			targetStr = fmt.Sprintf(" for target '%s'", params.Target)
		}
		logMsg := fmt.Sprintf("Executed script '%s'%s [%s, duration: %dms, sha256: %s].",
			filepath.Base(cleanScript), targetStr, statusStr, durationMs, codeHash[:12])
		_ = appendLogEntry(bDir, logMsg)
	}

	return res, nil
}
