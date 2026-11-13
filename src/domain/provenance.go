package domain

// ProvenanceEvent represents a single language-agnostic telemetry/lineage record
// emitted by an ETL script or notebook via OKF_PROVENANCE_OUT (JSONL).
type ProvenanceEvent struct {
	Type         string                 `json:"type"` // "input", "step", "output", "quality", "metric", "note", "drift"
	Timestamp    string                 `json:"timestamp,omitempty"`
	Step         string                 `json:"step,omitempty"`
	Tool         string                 `json:"tool,omitempty"`
	URI          string                 `json:"uri,omitempty"`
	Path         string                 `json:"path,omitempty"`
	Resource     string                 `json:"resource,omitempty"`
	Format       string                 `json:"format,omitempty"`
	HashSHA256   string                 `json:"hash_sha256,omitempty"`
	BytesCount   int64                  `json:"bytes_count,omitempty"`
	RowCount     int64                  `json:"row_count,omitempty"`
	Completeness float64                `json:"completeness,omitempty"`
	DurationMs   int64                  `json:"duration_ms,omitempty"`
	Status       string                 `json:"status,omitempty"`
	Metrics      map[string]interface{} `json:"metrics,omitempty"`
	Params       map[string]interface{} `json:"params,omitempty"`
	Fields       []ResourceField        `json:"fields,omitempty"`
	Message      string                 `json:"message,omitempty"`
	Severity     string                 `json:"severity,omitempty"`
}

// DriftAlert captures a schema or quality discrepancy between an automated
// Distribution instance and a curated Dataset contract without mutating curated knowledge.
type DriftAlert struct {
	Resource string `json:"resource"`
	Kind     string `json:"kind"` // "missing_column", "type_changed", "quality_threshold", "custom"
	Field    string `json:"field,omitempty"`
	Expected string `json:"expected,omitempty"`
	Actual   string `json:"actual,omitempty"`
	Severity string `json:"severity"` // "WARNING" | "ERROR"
	Message  string `json:"message"`
}

// ProvenanceReport aggregates all structured events collected during a script or notebook run.
type ProvenanceReport struct {
	Inputs  []ProvenanceEvent `json:"inputs,omitempty"`
	Steps   []ProvenanceEvent `json:"steps,omitempty"`
	Outputs []ProvenanceEvent `json:"outputs,omitempty"`
	Quality []ProvenanceEvent `json:"quality,omitempty"`
	Metrics []ProvenanceEvent `json:"metrics,omitempty"`
	Notes   []ProvenanceEvent `json:"notes,omitempty"`
	Drifts  []DriftAlert      `json:"drifts,omitempty"`
}
