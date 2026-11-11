# Tabular Query Engine & Registry Adapter (`src/adapters/outbound/engine`)

## Overview & Role

Provides the pure-Go fallback tabular query engine (`InMemTabularQueryEngine` in [`csv_query_engine.go`](./csv_query_engine.go)) implementing [`outbound.QueryEnginePort`](../../../ports/outbound/README.md) over CSV/TSV/TXT and `.zip`-archived resources, and `QueryEngineRegistry` ([`registry.go`](./registry.go)) implementing [`outbound.QueryEngineRegistryPort`](../../../ports/outbound/README.md).

## Interface

- **`InMemTabularQueryEngine` ([`csv_query_engine.go`](./csv_query_engine.go))**:
  - `[csv_query_engine.go:NewInMemTabularQueryEngine]`: `NewInMemTabularQueryEngine() *InMemTabularQueryEngine`
  - `[csv_query_engine.go:InMemTabularQueryEngine.Name]`: Returns `"inmem"`.
  - `[csv_query_engine.go:InMemTabularQueryEngine.SupportsFormat]`: Supports `"csv"`, `"tsv"`, `"txt"`.
  - `[csv_query_engine.go:InMemTabularQueryEngine.SupportsCrossFormatJoin]`: Returns `false`.
  - `[csv_query_engine.go:InMemTabularQueryEngine.RegisterTable]`: No-op (`nil`).
  - `[csv_query_engine.go:InMemTabularQueryEngine.Execute]`: Streams CSV/TSV rows, applies case-insensitive column projections (`req.Select`), equality filters (`req.Filters`), and row limits (`req.Limit`).
  - `[csv_query_engine.go:InMemTabularQueryEngine.ExecuteSQL]`: Parses `domain.SimpleSelect`, resolves a single local table binding, transparently extracts `.zip` archives (with optional RFC 3986 `#fragment` subresource selection), and delegates to `Execute`.
- **`QueryEngineRegistry` ([`registry.go`](./registry.go))**:
  - `[registry.go:NewQueryEngineRegistry]`: `NewQueryEngineRegistry(defaultEngines ...outbound.QueryEnginePort) *QueryEngineRegistry`
  - `[registry.go:QueryEngineRegistry.RegisterEngine]`: `RegisterEngine(engine outbound.QueryEnginePort)`
  - `[registry.go:QueryEngineRegistry.GetEngine]`: `GetEngine(name string) (outbound.QueryEnginePort, bool)`
  - `[registry.go:QueryEngineRegistry.SelectEngine]`: `SelectEngine(preferred string, formats []string) (outbound.QueryEnginePort, error)`

## Invariants

- **Parquet Guard (`[csv_query_engine.go:InMemTabularQueryEngine.Execute]`)**: Rejects `"parquet"` format or `PAR1` magic bytes with an explicit error directing callers to a columnar engine (e.g., DuckDB).
- **Schema Typo Assistance (`[csv_query_engine.go:suggestClosestStrings]`)**: Missing projected columns fail with top-3 closest header matches ranked by Levenshtein edit distance.
- **Single-Table & Local-Only SQL (`[csv_query_engine.go:InMemTabularQueryEngine.ExecuteSQL]`)**: Rejects unsupported SQL syntax (`domain.ParseSimpleSelect`), remote `http(s)://` physical paths, and multi-table bindings (`selectSingleBinding`).
- **Thread-Safe Deterministic Selection (`[registry.go:QueryEngineRegistry.SelectEngine]`)**: Guards map access with `sync.RWMutex` and sorts engine names alphabetically (`orderedEnginesLocked`) before capability matching.

## Side Effects

- **Local File & Archive Reads (`[csv_query_engine.go:InMemTabularQueryEngine.ExecuteSQL]`)**: Reads local CSV/TSV/ZIP files from disk (`os.ReadFile`, `archive/zip.NewReader`).

## Verification

```bash
cd src && go test ./adapters/outbound/engine/...
```
