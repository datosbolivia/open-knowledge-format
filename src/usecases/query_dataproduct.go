package usecases

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/datosbolivia/open-knowledge-format/src/domain"
	"github.com/datosbolivia/open-knowledge-format/src/ports/outbound"
)

// QueryDataProductUseCase orchestrates fetching resource data and executing tabular and SQL queries.
type QueryDataProductUseCase struct {
	engine        outbound.QueryEnginePort
	storage       outbound.StoragePort
	triadResolver outbound.TriadResolverPort
	httpClient    *http.Client
}

// NewQueryDataProductUseCase initializes the query use case.
func NewQueryDataProductUseCase(
	engine outbound.QueryEnginePort,
	storage outbound.StoragePort,
	triadResolver outbound.TriadResolverPort,
) *QueryDataProductUseCase {
	return &QueryDataProductUseCase{
		engine:        engine,
		storage:       storage,
		triadResolver: triadResolver,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}
}

// ExecuteSQL parses canonical triads from a SQL statement, resolves physical paths, and delegates to the query engine.
func (uc *QueryDataProductUseCase) ExecuteSQL(ctx context.Context, sqlQuery string) (*domain.QueryResult, error) {
	return uc.ExecuteSQLWithOptions(ctx, sqlQuery, domain.QueryOptions{})
}

// ExecuteSQLWithOptions executes a SQL query with explicit options (engine choice, table overrides, timeouts).
func (uc *QueryDataProductUseCase) ExecuteSQLWithOptions(
	ctx context.Context,
	sqlQuery string,
	opts domain.QueryOptions,
) (*domain.QueryResult, error) {
	triads := domain.ExtractTriadsFromSQL(sqlQuery)
	resolvedMap := make(map[string]domain.ResolvedResource)
	bindings := make(map[string]domain.TableBinding)

	// 1. Apply custom table mappings from options
	for name, path := range opts.TableMapping {
		format := "csv"
		lower := strings.ToLower(path)
		if strings.HasSuffix(lower, ".parquet") || strings.HasSuffix(lower, ".pq") {
			format = "parquet"
		} else if strings.HasSuffix(lower, ".json") || strings.HasSuffix(lower, ".jsonl") {
			format = "json"
		} else if strings.HasSuffix(lower, ".tsv") {
			format = "tsv"
		}
		bindings[name] = domain.TableBinding{
			Name:         name,
			PhysicalPath: path,
			Format:       format,
		}
	}

	// 2. Resolve triads referenced in query
	for _, t := range triads {
		key := t.String()
		if _, exists := bindings[key]; exists {
			continue
		}
		if uc.triadResolver != nil {
			res, err := uc.triadResolver.ResolveTriad(ctx, t)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve triad '%s': %w", key, err)
			}
			resolvedMap[key] = *res
			bindings[key] = domain.TableBinding{
				Name:         key,
				PhysicalPath: res.PhysicalURI,
				Format:       res.Format,
			}
		} else {
			resolvedMap[key] = domain.ResolvedResource{
				Triad:       t,
				PhysicalURI: t.Resource,
				Format:      "csv",
			}
			bindings[key] = domain.TableBinding{
				Name:         key,
				PhysicalPath: t.Resource,
				Format:       "csv",
			}
		}
	}

	req := domain.SQLQueryRequest{
		SQLQuery:       sqlQuery,
		Triads:         triads,
		ResolvedTables: resolvedMap,
		Bindings:       bindings,
		Options:        opts,
	}

	return uc.engine.ExecuteSQL(ctx, req)
}

// Query executes a tabular query against a target resource URI.
func (uc *QueryDataProductUseCase) Query(ctx context.Context, req domain.QueryRequest) (*domain.QueryResult, error) {
	if strings.TrimSpace(req.ResourceURI) == "" {
		return nil, fmt.Errorf("resource URI cannot be empty")
	}

	data, format, err := uc.loadResourceData(ctx, req.ResourceURI)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resource data: %w", err)
	}

	return uc.engine.Execute(ctx, req, data, format)
}

func (uc *QueryDataProductUseCase) loadResourceData(ctx context.Context, uri string) ([]byte, string, error) {
	baseURI := uri
	fragment := ""
	if hashIdx := strings.Index(uri, "#"); hashIdx != -1 {
		baseURI = strings.TrimSpace(uri[:hashIdx])
		fragment = strings.TrimPrefix(strings.TrimSpace(uri[hashIdx+1:]), "/")
	}

	format := strings.TrimPrefix(strings.ToLower(filepath.Ext(baseURI)), ".")
	if format == "" {
		format = "csv"
	}

	rawBytes, err := uc.fetchRawBytes(ctx, baseURI)
	if err != nil {
		return nil, "", err
	}

	// Transparent composite ZIP archive extraction (Zero-Friction Archive Mounting)
	if format == "zip" || (len(rawBytes) >= 4 && rawBytes[0] == 'P' && rawBytes[1] == 'K' && rawBytes[2] == 0x03 && rawBytes[3] == 0x04) {
		return extractFromZipBytes(rawBytes, fragment)
	}

	return rawBytes, format, nil
}

func (uc *QueryDataProductUseCase) fetchRawBytes(ctx context.Context, uri string) ([]byte, error) {
	// 1. Try cached storage if available
	if uc.storage != nil {
		if exists, _ := uc.storage.Exists(ctx, uri); exists {
			data, err := uc.storage.Load(ctx, uri)
			if err == nil {
				return data, nil
			}
		}
	}

	// 2. Local file
	if strings.HasPrefix(uri, "file://") || (!strings.HasPrefix(uri, "http://") && !strings.HasPrefix(uri, "https://")) {
		localPath := strings.TrimPrefix(uri, "file://")
		return os.ReadFile(localPath)
	}

	// 3. Remote HTTP fetch
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	resp, err := uc.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Cache to storage if available
	if uc.storage != nil {
		_ = uc.storage.Save(ctx, uri, body)
	}

	return body, nil
}

// extractFromZipBytes extracts a specific #fragment subfile or the first tabular file from a ZIP archive in memory.
func extractFromZipBytes(zipData []byte, fragment string) ([]byte, string, error) {
	zr, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, "", fmt.Errorf("failed to open ZIP container: %w", err)
	}

	var selected *zip.File
	cleanFrag := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(fragment), "#"))

	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		fName := strings.ToLower(strings.TrimPrefix(f.Name, "/"))
		fBase := strings.ToLower(filepath.Base(f.Name))
		if cleanFrag != "" {
			if fName == cleanFrag || fBase == cleanFrag || strings.HasSuffix(fName, "/"+cleanFrag) {
				selected = f
				break
			}
		} else {
			ext := strings.ToLower(filepath.Ext(f.Name))
			if ext == ".csv" || ext == ".tsv" || ext == ".parquet" || ext == ".json" {
				selected = f
				break
			}
		}
	}

	if selected == nil && len(zr.File) > 0 && cleanFrag == "" {
		for _, f := range zr.File {
			if !f.FileInfo().IsDir() {
				selected = f
				break
			}
		}
	}

	if selected == nil {
		available := make([]string, 0, len(zr.File))
		for _, f := range zr.File {
			if !f.FileInfo().IsDir() {
				available = append(available, f.Name)
			}
		}
		return nil, "", fmt.Errorf("subresource %q not found inside ZIP container; available internal files: %v", fragment, available)
	}

	rc, err := selected.Open()
	if err != nil {
		return nil, "", fmt.Errorf("failed to open internal ZIP file %q: %w", selected.Name, err)
	}
	defer rc.Close()

	extracted, err := io.ReadAll(rc)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read internal ZIP file %q: %w", selected.Name, err)
	}

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(selected.Name)), ".")
	if ext == "" {
		ext = "csv"
	}
	return extracted, ext, nil
}
