# Outbound Adapters (`src/adapters/outbound`)

## Overview & Role

`src/adapters/outbound` is an Intermediate Blueprint (Hay) containing secondary driven adapters implementing the outbound ports ([`src/ports/outbound`](../../ports/outbound/README.md)) for configuration loading, in-memory SQL/tabular query execution, multi-standard manifest parsing, catalog/node resolution, and sandboxed local file storage.

## Subtree Index

| Subpackage | Implemented Port(s) | Key Types |
| :--- | :--- | :--- |
| [`config/`](./config/README.md) | `outbound.ConfigLoaderPort` | `FileAndEnvConfigLoader` |
| [`engine/`](./engine/README.md) | `outbound.QueryEnginePort`, `outbound.QueryEngineRegistryPort` | `InMemTabularQueryEngine`, `QueryEngineRegistry` |
| [`manifests/`](./manifests/README.md) | `outbound.UnifiedMetadataReaderPort`, `outbound.TriadResolverPort`, `outbound.ManifestParserPort` | `UnifiedMetadataReader`, `DataPackageJSONParser`, `DataPackageYAMLParser`, `OKFMarkdownParser`, `LLMSTxtParser` |
| [`resolvers/`](./resolvers/README.md) | `outbound.CatalogResolverPort`, `outbound.DataProductResolverPort` | `LLMSTxtCatalogResolver`, `NodeDataProductResolver` |
| [`storage/`](./storage/README.md) | `outbound.StoragePort` | `FileStorage` |

## Interface

- **Configuration ([`./config/README.md`](./config/README.md))**: `[loader.go:NewConfigLoader]`, `[loader.go:FileAndEnvConfigLoader.Load]`.
- **Query Engine ([`./engine/README.md`](./engine/README.md))**: `[csv_query_engine.go:NewInMemTabularQueryEngine]`, `[registry.go:NewQueryEngineRegistry]`.
- **Manifests & Contract Synthesis ([`./manifests/README.md`](./manifests/README.md))**: `[unified_reader.go:NewUnifiedMetadataReader]`, `[datapackage_json_parser.go:NewDataPackageJSONParser]`, `[datapackage_yaml_parser.go:NewDataPackageYAMLParser]`, `[okf_markdown_parser.go:NewOKFMarkdownParser]`, `[llmstxt_parser.go:NewLLMSTxtParser]`, `[openapi_and_cli_synthesizer.go:SynthesizeContractMarkdown]`.
- **Resolvers ([`./resolvers/README.md`](./resolvers/README.md))**: `[llmstxt_resolver.go:NewLLMSTxtCatalogResolver]`, `[node_resolver.go:NewNodeDataProductResolver]`.
- **Storage ([`./storage/README.md`](./storage/README.md))**: `[file_storage.go:NewFileStorage]`.

## Invariants

- **Hexagonal Dependency Rule**: Outbound adapters depend only on [`src/domain`](../../domain/README.md) and [`src/ports/outbound`](../../ports/outbound/README.md).
- **Path & Format Safety**: [`./storage/README.md`](./storage/README.md) enforces `baseDir` confinement and atomic `.tmp` writes; [`./engine/README.md`](./engine/README.md) rejects unsupported SQL/Parquet inputs explicitly; [`./manifests/README.md`](./manifests/README.md) preserves non-`index.md` concept metadata and prevents non-DataPackage contracts from eclipsing `datapackage.yaml`.

## Side Effects

- Aggregates environment variable reads, local filesystem reads/writes, `.zip` archive extraction, and outbound HTTP `GET` requests across child packages.

## Verification

```bash
cd src && go test ./adapters/outbound/...
```
