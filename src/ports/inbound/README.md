# `inbound` Package (`src/ports/inbound`)

## Overview & Role

Package `inbound` ([`services.go`](./services.go)) defines the primary driving (inbound) service ports for the Open Knowledge Format (OKF / ODKF v0.2) hexagonal architecture. Inbound adapters—including the CLI ([`src/adapters/inbound/cli`](../../adapters/inbound/cli/README.md)), C-shared FFI exports ([`src/adapters/inbound/cshared`](../../adapters/inbound/cshared/README.md)), and WebAssembly entrypoint ([`src/adapters/inbound/wasm`](../../adapters/inbound/wasm/README.md))—invoke application use cases ([`src/usecases`](../../usecases/README.md)) through these abstract service contracts.

## Interface

- **`[services.go:CatalogServicePort]`**
  - `DiscoverAll(ctx context.Context) (*domain.Catalog, error)`: Discovers and aggregates all configured sovereign catalogs.
  - `Discover(ctx context.Context, catalogURL string) (*domain.Catalog, error)`: Fetches and resolves a specific catalog endpoint (`llms.txt`, `index.md`, or federated catalog).
  - `Search(ctx context.Context, catalogURL, keyword string) ([]domain.CatalogEntry, error)`: Searches catalog entries by keyword across the target or default catalog.
- **`[services.go:DataProductServicePort]`**
  - `Resolve(ctx context.Context, uriOrURL string) (*domain.DataProduct, error)`: Resolves a canonical triad or URL into a hydrated `domain.DataProduct`.
  - `Validate(ctx context.Context, dp *domain.DataProduct) error`: Validates a `DataProduct` against ODKF producer/consumer invariants.
- **`[services.go:QueryServicePort]`**
  - `Query(ctx context.Context, req domain.QueryRequest) (*domain.QueryResult, error)`: Executes a structured tabular query request against a data resource.
  - `ExecuteSQL(ctx context.Context, sqlQuery string) (*domain.QueryResult, error)`: Executes an ANSI SQL query (supporting canonical `catalog:dataset:resource` triads) against resolved resources.
- **`[services.go:ConfigServicePort]`**
  - `GetConfig() domain.Config`: Returns the active runtime configuration snapshot.
- **`[services.go:PublicationServicePort]`**
  - `Publish(ctx context.Context, datasetPath string, targets []string, options map[string]interface{}) ([]domain.PublicationTargetResult, error)`: Publishes an OKF bundle or dataset across one or more target platforms.
- **`[services.go:SemanticServicePort]`**
  - `AlignSemantics(ctx context.Context, manifest *domain.PackageManifest) ([]domain.SemanticFieldMapping, error)`: Computes semantic field mappings (`ValueLabel` -> `Concept` -> `SKOS`) across a package manifest.
  - `GenerateCaseExpression(ctx context.Context, mapping domain.SemanticFieldMapping, tablePrefix string) string`: Compiles a `SemanticFieldMapping` into a deterministic ANSI SQL `CASE ... END` projection.
- **`[services.go:HarvesterServicePort]`**
  - `DiscoverWellKnown(ctx context.Context, targetURL string) (*domain.WellKnownDiscovery, error)`: Probes `/.well-known/datamesh.json` capability metadata on a remote host.
  - `HarvestCKAN(ctx context.Context, ckanBaseURL string, options map[string]interface{}) (*domain.CKANHarvestResult, error)`: Harvests packages from a CKAN Action API v3 portal.
  - `ReconstructOKF(ctx context.Context, pkg *domain.CKANPackage) (*domain.PackageManifest, error)`: Reconstructs a canonical OKF `PackageManifest` from a raw `CKANPackage`.
- **`[services.go:LineageServicePort]`**
  - `BuildGraph(ctx context.Context, searchDirs []string, catalogURL string) (*domain.LineageGraph, error)`: Builds the dataset provenance and derivation DAG across local directories and catalogs.
  - `GetLineage(ctx context.Context, datasetIDOrPath string, searchDirs []string, catalogURL string) (map[string]interface{}, error)`: Retrieves upstream/downstream lineage metadata for a specific dataset or path.
- **`[services.go:InteroperabilityServicePort]`**
  - `Ingest(ctx context.Context, target string, authType string) (*domain.DataProduct, error)`: Ingests heterogeneous external metadata as-is into a canonical ODKF `DataProduct`.
  - `ExportCKANDataPackage(ctx context.Context, dp *domain.DataProduct) ([]byte, error)`: Exports an ODKF `DataProduct` to CKAN/Frictionless `datapackage.json` format while preserving ODKF extension keys.
  - `ExportDCAT(ctx context.Context, dp *domain.DataProduct) ([]byte, error)`: Exports an ODKF `DataProduct` to W3C DCAT JSON-LD format.

## Invariants

- **Pure Interface Boundary (`[services.go:CatalogServicePort]`–`[services.go:InteroperabilityServicePort]`)**: Contains strictly Go `interface` declarations and zero implementation logic.
- **Dependency Direction**: Depends exclusively on `context` and [`github.com/datosbolivia/open-knowledge-format/src/domain`](../../domain/README.md). Never imports `src/usecases` or `src/adapters`.
- **Context Propagation**: All I/O-capable operations accept `context.Context` as their first parameter to enforce cancellation and timeout propagation from driving adapters.

## Side Effects

- None (`["None"]` — pure interface declarations).

## Verification

```bash
cd src && go test ./...
```
