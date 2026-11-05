package engine

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/datosbolivia/open-knowledge-format/src/domain"
)

// InMemTabularQueryEngine executes filtering and projection queries over CSV and delimited data.
type InMemTabularQueryEngine struct{}

// NewInMemTabularQueryEngine creates an instance of the tabular query engine.
func NewInMemTabularQueryEngine() *InMemTabularQueryEngine {
	return &InMemTabularQueryEngine{}
}

func (e *InMemTabularQueryEngine) Name() string {
	return "inmem"
}

func (e *InMemTabularQueryEngine) SupportsFormat(format string) bool {
	f := strings.ToLower(strings.TrimSpace(format))
	return f == "csv" || f == "tsv" || f == "txt"
}

func (e *InMemTabularQueryEngine) SupportsCrossFormatJoin() bool {
	return false
}

func (e *InMemTabularQueryEngine) RegisterTable(ctx context.Context, table domain.TableBinding) error {
	return nil
}

// Execute parses raw delimited data, applies column selection, filters, and limits.
func (e *InMemTabularQueryEngine) Execute(
	ctx context.Context,
	req domain.QueryRequest,
	rawData []byte,
	format string,
) (*domain.QueryResult, error) {
	if strings.EqualFold(format, "parquet") || (len(rawData) >= 4 && string(rawData[:4]) == "PAR1") {
		return nil, fmt.Errorf("binary Parquet format requires a columnar query engine (e.g. DuckDB)")
	}
	start := time.Now()

	reader := csv.NewReader(bytes.NewReader(rawData))
	if strings.ToLower(format) == "tsv" {
		reader.Comma = '\t'
	}

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read table headers: %w", err)
	}

	colIndex := make(map[string]int)
	for i, h := range headers {
		colIndex[strings.ToLower(strings.TrimSpace(h))] = i
	}

	outHeaders, projectIdx, projErr := projectColumns(headers, colIndex, req.Select)
	if projErr != nil {
		return nil, projErr
	}

	// Filter mapping (case-insensitive column names)
	filterIndices := make(map[int]string)
	for k, v := range req.Filters {
		normKey := strings.ToLower(strings.TrimSpace(k))
		if idx, found := colIndex[normKey]; found {
			filterIndices[idx] = strings.ToLower(strings.TrimSpace(v))
		}
	}

	var matchedRows [][]interface{}

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue // skip malformed lines
		}

		// Apply filters
		matches := true
		for idx, expectedVal := range filterIndices {
			if idx >= len(record) || strings.ToLower(strings.TrimSpace(record[idx])) != expectedVal {
				matches = false
				break
			}
		}

		if matches {
			if len(projectIdx) > 0 {
				row := make([]interface{}, len(projectIdx))
				for i, idx := range projectIdx {
					if idx >= 0 && idx < len(record) {
						row[i] = record[idx]
					} else {
						row[i] = ""
					}
				}
				matchedRows = append(matchedRows, row)
			} else {
				row := make([]interface{}, len(record))
				for i, val := range record {
					row[i] = val
				}
				matchedRows = append(matchedRows, row)
			}

			if req.Limit > 0 && len(matchedRows) >= req.Limit {
				break
			}
		}
	}

	elapsed := time.Since(start)
	return domain.NewQueryResult(outHeaders, matchedRows, elapsed), nil
}

// projectColumns resolves the requested column projection to output headers and
// source indices. An empty selection means "all columns".
func projectColumns(headers []string, colIndex map[string]int, selectCols []string) ([]string, []int, error) {
	if len(selectCols) == 0 {
		return headers, nil, nil
	}
	outHeaders := make([]string, 0, len(selectCols))
	projectIdx := make([]int, 0, len(selectCols))
	for _, col := range selectCols {
		idx, found := colIndex[strings.ToLower(strings.TrimSpace(col))]
		if !found {
			suggestions := suggestClosestStrings(col, headers, 3)
			return nil, nil, fmt.Errorf("column %q not found in table; closest matches: %v (available columns: %v)", col, suggestions, headers)
		}
		outHeaders = append(outHeaders, headers[idx])
		projectIdx = append(projectIdx, idx)
	}
	return outHeaders, projectIdx, nil
}

// suggestClosestStrings ranks candidates by Levenshtein edit distance to guide LLM agents on schema typos.
func suggestClosestStrings(target string, candidates []string, maxResults int) []string {
	tNorm := strings.ToLower(strings.TrimSpace(target))
	type scored struct {
		val  string
		dist int
	}
	var list []scored
	for _, c := range candidates {
		cNorm := strings.ToLower(strings.TrimSpace(c))
		list = append(list, scored{val: c, dist: levenshteinDistance(tNorm, cNorm)})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].dist == list[j].dist {
			return list[i].val < list[j].val
		}
		return list[i].dist < list[j].dist
	})
	if maxResults > len(list) {
		maxResults = len(list)
	}
	out := make([]string, 0, maxResults)
	for i := 0; i < maxResults; i++ {
		out = append(out, list[i].val)
	}
	return out
}

func levenshteinDistance(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)
	la, lb := len(ra), len(rb)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	dp := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		dp[j] = j
	}
	for i := 1; i <= la; i++ {
		prev := dp[0]
		dp[0] = i
		for j := 1; j <= lb; j++ {
			temp := dp[j]
			cost := 0
			if ra[i-1] != rb[j-1] {
				cost = 1
			}
			del := dp[j] + 1
			ins := dp[j-1] + 1
			sub := prev + cost
			minVal := del
			if ins < minVal {
				minVal = ins
			}
			if sub < minVal {
				minVal = sub
			}
			dp[j] = minVal
			prev = temp
		}
	}
	return dp[lb]
}

// ExecuteSQL executes the supported SQL subset (see domain.SimpleSelect) over a
// single resolved CSV/TSV resource (including transparent extraction from local .zip archives).
// Queries outside that subset are rejected explicitly instead of silently returning wrong rows;
// callers needing full ANSI SQL must use the DuckDB engine.
func (e *InMemTabularQueryEngine) ExecuteSQL(ctx context.Context, req domain.SQLQueryRequest) (*domain.QueryResult, error) {
	selectStmt, err := domain.ParseSimpleSelect(req.SQLQuery)
	if err != nil {
		return nil, err
	}

	binding, err := selectSingleBinding(req, selectStmt.Table)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(binding.PhysicalPath, "http://") || strings.HasPrefix(binding.PhysicalPath, "https://") {
		return nil, fmt.Errorf("pure-Go fallback engine cannot read remote resource %q; use the DuckDB engine", binding.PhysicalPath)
	}

	localPath := strings.TrimPrefix(binding.PhysicalPath, "file://")
	fragment := ""
	if hashIdx := strings.Index(localPath, "#"); hashIdx != -1 {
		fragment = strings.TrimSpace(localPath[hashIdx+1:])
		localPath = strings.TrimSpace(localPath[:hashIdx])
	}

	rawData, err := os.ReadFile(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read table '%s' at %s: %w", binding.Name, binding.PhysicalPath, err)
	}

	effectiveFormat := binding.Format
	if strings.EqualFold(filepath.Ext(localPath), ".zip") || (len(rawData) >= 4 && rawData[0] == 'P' && rawData[1] == 'K' && rawData[2] == 0x03 && rawData[3] == 0x04) {
		extracted, extFmt, zipErr := extractZipEntryInMem(rawData, fragment)
		if zipErr != nil {
			return nil, zipErr
		}
		rawData = extracted
		effectiveFormat = extFmt
	}

	return e.Execute(ctx, domain.QueryRequest{
		ResourceURI: binding.PhysicalPath,
		Query:       req.SQLQuery,
		Select:      selectStmt.Columns,
		Filters:     selectStmt.Where,
		Limit:       selectStmt.Limit,
	}, rawData, effectiveFormat)
}

func extractZipEntryInMem(zipData []byte, fragment string) ([]byte, string, error) {
	zr, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, "", fmt.Errorf("failed to open ZIP archive: %w", err)
	}
	cleanFrag := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(fragment), "#"))
	var selected *zip.File
	var names []string
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		names = append(names, f.Name)
		fLower := strings.ToLower(strings.TrimPrefix(f.Name, "/"))
		if cleanFrag != "" {
			if fLower == cleanFrag || strings.ToLower(filepath.Base(f.Name)) == cleanFrag {
				selected = f
				break
			}
		} else if selected == nil {
			selected = f
		}
	}
	if selected == nil {
		return nil, "", fmt.Errorf("subresource %q not found inside ZIP archive; available files: %v", fragment, names)
	}
	rc, err := selected.Open()
	if err != nil {
		return nil, "", err
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, "", err
	}
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(selected.Name)), ".")
	if ext == "" {
		ext = "csv"
	}
	return data, ext, nil
}

// selectSingleBinding resolves the single table referenced by the query from the
// resolved triads or explicit mappings, rejecting multi-table queries instead of
// picking an arbitrary table (map iteration order is not deterministic).
func selectSingleBinding(req domain.SQLQueryRequest, tableName string) (domain.TableBinding, error) {
	if tableName != "" {
		if binding, ok := req.Bindings[tableName]; ok {
			return binding, nil
		}
		if res, ok := req.ResolvedTables[tableName]; ok {
			return bindingFromResolved(tableName, res), nil
		}
	}

	total := len(req.Bindings) + len(req.ResolvedTables)
	switch {
	case total == 0:
		return domain.TableBinding{}, fmt.Errorf("no table resolved for query %q", req.SQLQuery)
	case total == 1:
		for _, binding := range req.Bindings {
			return binding, nil
		}
		for key, res := range req.ResolvedTables {
			return bindingFromResolved(key, res), nil
		}
	}
	return domain.TableBinding{}, fmt.Errorf("pure-Go fallback engine supports a single table per query (use the DuckDB engine for JOIN / multi-table queries)")
}

func bindingFromResolved(name string, res domain.ResolvedResource) domain.TableBinding {
	return domain.TableBinding{Name: name, PhysicalPath: res.PhysicalURI, Format: res.Format}
}
