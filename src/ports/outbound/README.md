# `outbound` Package (`src/ports/outbound`)

## Overview & Role

Package `outbound` defines the driven (secondary/outbound) infrastructure ports for the Open Knowledge Format (OKF / ODKF v0.2) hexagonal architecture ([`SPEC.md`](../../../SPEC.md) §16.1–16.2). Application use cases ([`src/usecases`](../../usecases/README.md)) depend on these interfaces, while concrete outbound adapters ([`src/adapters/outbound`](../../adapters/outbound/README.md)) implement them for configuration loading, local/remote storage, manifest parsing, catalog/node resolution, publishing, and tabular/SQL query execution. Non-standard platforms (such as CKAN) are handled as external services rather than core ports.

## Interface

### Configuration ([`config.go`](./config.go))
- **`[config.go:ConfigLoaderPort]`**
  - `Load(explicitFilePath string) (domain.Config, error)`: Loads runtime configuration applying precedence: Environment Variables > Config File (`datamesh.json`/`yaml`/`yml`) > Default values.

### Discovery & Authentication ([`harvester.go`](./harvester.go))
- **`[harvester.go:AuthProviderPort]`**
  - `GetAuthHeaders(ctx context.Context) (map[string]string, error)`: Produces HTTP authentication headers for outbound requests.
  - `GetAuthType() string`: Returns the authentication scheme identifier (`none`, `api_key`, `keycloak`, `oauth2`).
- **`[harvester.go:WellKnownResolverPort]`**
  - `FetchWellKnown(ctx context.Context, targetURL string) (*domain.WellKnownDiscovery, error)`: Fetches sovereign capability discovery metadata from a target URL.

### Package & Manifest Readers ([`package_reader.go`](./package_reader.go))
- **`[package_reader.go:ManifestParserPort]`**
  - `Name() string`: Unique parser identifier.
  - `CanParse(filenameOrURL string, data []byte) bool`: Determines whether the parser supports the given resource path or byte payload.
  - `ParseManifest(ctx context.Context, data []byte, baseURL string) (*domain.PackageManifest, error)`: Parses raw descriptor bytes (`datapackage.json`, `datapackage.yaml`, OKF `index.md`, `llms.txt`) into a normalized `domain.PackageManifest`.
- **`[package_reader.go:UnifiedMetadataReaderPort]`**
  - `RegisterParser(parser ManifestParserPort)`: Registers an extensible manifest parser plugin.
  - `ReadPackage(ctx context.Context, uriOrPath string) (*domain.PackageManifest, error)`: Reads and parses a package manifest from a local path or remote URL.
  - `ReadNode(ctx context.Context, uriOrPath string) (*domain.DataProduct, error)`: Reads an OKF node descriptor into a `domain.DataProduct`.
  - `ReadCatalog(ctx context.Context, uriOrPath string) (*domain.Catalog, error)`: Reads a catalog index into a `domain.Catalog`.

### Publication ([`publisher.go`](./publisher.go))
- **`[publisher.go:PublisherTargetPort]`**
  - `Name() string`: Target platform identifier.
  - `CanPublish(target string) bool`: Reports whether this adapter handles the requested publication target.
  - `Publish(ctx context.Context, manifest *domain.PackageManifest, datasetPath string, options map[string]interface{}) (domain.PublicationTargetResult, error)`: Publishes a dataset and its manifest to the external platform.

### Query Engines ([`query.go`](./query.go))
- **`[query.go:QueryEnginePort]`**
  - `Name() string`: Engine identifier (e.g., `inmem`, `duckdb`).
  - `SupportsFormat(format string) bool`: Reports whether the engine natively queries the given data format (`csv`, `parquet`, etc.).
  - `SupportsCrossFormatJoin() bool`: Reports whether the engine supports SQL joins across heterogeneous file formats.
  - `Execute(ctx context.Context, req domain.QueryRequest, rawData []byte, format string) (*domain.QueryResult, error)`: Executes a structured `QueryRequest` over raw resource bytes.
  - `ExecuteSQL(ctx context.Context, req domain.SQLQueryRequest) (*domain.QueryResult, error)`: Executes a `SQLQueryRequest` across registered table bindings.
  - `RegisterTable(ctx context.Context, table domain.TableBinding) error`: Binds a logical table alias or canonical triad to a physical resource inside the engine.
- **`[query.go:QueryEngineRegistryPort]`**
  - `RegisterEngine(engine QueryEnginePort)`: Registers a query engine implementation.
  - `GetEngine(name string) (QueryEnginePort, bool)`: Looks up a registered engine by name.
  - `SelectEngine(preferred string, formats []string) (QueryEnginePort, error)`: Selects a compatible query engine honoring the preferred engine name and required resource formats.

### Catalog, Data Product & Triad Resolvers ([`resolvers.go`](./resolvers.go), [`triad_resolver.go`](./triad_resolver.go))
- **`[resolvers.go:CatalogResolverPort]`**
  - `FetchCatalog(ctx context.Context, url string) (*domain.Catalog, error)`: Fetches and parses a remote or local catalog endpoint (`llms.txt`, `index.md`).
- **`[resolvers.go:DataProductResolverPort]`**
  - `FetchDataProduct(ctx context.Context, resolvedURL string) (*domain.DataProduct, error)`: Fetches a remote or local OKF node and extracts its `DataProduct` aggregate.
- **`[triad_resolver.go:TriadResolverPort]`**
  - `ResolveTriad(ctx context.Context, triad domain.ResourceTriad) (*domain.ResolvedResource, error)`: Resolves a canonical `catalog:dataset:resource` triad into a physical URI and data format.

### Storage ([`storage.go`](./storage.go))
- **`[storage.go:StoragePort]`**
  - `Save(ctx context.Context, key string, data []byte) error`: Persists binary or text payload under `key`.
  - `Load(ctx context.Context, key string) ([]byte, error)`: Loads cached payload bytes for `key`.
  - `Exists(ctx context.Context, key string) (bool, error)`: Checks whether `key` exists in storage.

## Invariants

- **Dependency Inversion Boundary**: Depends only on standard library `context` and [`github.com/datosbolivia/open-knowledge-format/src/domain`](../../domain/README.md). Never imports concrete adapters (`src/adapters/*`) or use cases (`src/usecases`).
- **Side-Effect Isolation**: All filesystem, network, and third-party API side effects are isolated behind these interfaces so [`src/usecases`](../../usecases/README.md) remains deterministic and unit-testable with in-memory mocks.

## Side Effects

- None (`["None"]` — pure interface declarations).

## Verification

```bash
cd src && go test ./...
```
