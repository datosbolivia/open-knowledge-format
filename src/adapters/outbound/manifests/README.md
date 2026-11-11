# Unified Metadata & Manifest Parsers (`src/adapters/outbound/manifests`)

## Overview & Role

Implements [`outbound.UnifiedMetadataReaderPort`](../../../ports/outbound/README.md), [`outbound.TriadResolverPort`](../../../ports/outbound/README.md), and format-specific [`outbound.ManifestParserPort`](../../../ports/outbound/README.md) adapters for Frictionless Data Package (`datapackage.json`, `datapackage.yaml`), OKF v0.2 Markdown (`index.md` / concept documents), OpenAPI (`openapi.yaml`), `owl-cli`/`okf-cli` contracts, and federated `llms.txt` catalogs.

## Interface

- **`UnifiedMetadataReader` ([`unified_reader.go`](./unified_reader.go))**:
  - `[unified_reader.go:NewUnifiedMetadataReader]`: `NewUnifiedMetadataReader(catalogURL string, timeout time.Duration) *UnifiedMetadataReader` (defaults timeout to `30s` if `<= 0`; registers JSON, YAML, OKF Markdown, and `llms.txt` parsers).
  - `[unified_reader.go:UnifiedMetadataReader.RegisterParser]`: `RegisterParser(parser outbound.ManifestParserPort)`
  - `[unified_reader.go:UnifiedMetadataReader.ReadPackage]`: `ReadPackage(ctx context.Context, uriOrPath string) (*domain.PackageManifest, error)`
  - `[unified_reader.go:UnifiedMetadataReader.ReadNode]`: `ReadNode(ctx context.Context, uriOrPath string) (*domain.DataProduct, error)` — Resolves local directories, `.md` concept files, slugs, or remote URLs; parses YAML frontmatter; hydrates external OpenAPI (`openapi.yaml`), `owl-cli`/`okf-cli`, and DataPackage contracts; strips `ojs` code blocks (`StripOJSBlocks`); and synthesizes Markdown contract tables (`SynthesizeContractMarkdown`).
  - `[unified_reader.go:UnifiedMetadataReader.ReadCatalog]`: `ReadCatalog(ctx context.Context, uriOrPath string) (*domain.Catalog, error)`
  - `[unified_reader.go:UnifiedMetadataReader.ResolveTriad]`: `ResolveTriad(ctx context.Context, triad domain.ResourceTriad) (*domain.ResolvedResource, error)` — Resolves canonical triads across local bundle directories and remote catalogs, preserving RFC 3986 `#fragment` subresource identifiers.
- **Parsers & Contract Synthesizers**:
  - **`DataPackageJSONParser` ([`datapackage_json_parser.go`](./datapackage_json_parser.go))**: `[datapackage_json_parser.go:NewDataPackageJSONParser]`, `[datapackage_json_parser.go:DataPackageJSONParser.Name]`, `[datapackage_json_parser.go:DataPackageJSONParser.CanParse]`, `[datapackage_json_parser.go:DataPackageJSONParser.ParseManifest]` — parses `datapackage.json` and expands `internal_resources` / `subresources` into `#fragment` resources.
  - **`DataPackageYAMLParser` ([`datapackage_yaml_parser.go`](./datapackage_yaml_parser.go))**: `[datapackage_yaml_parser.go:NewDataPackageYAMLParser]`, `[datapackage_yaml_parser.go:DataPackageYAMLParser.Name]`, `[datapackage_yaml_parser.go:DataPackageYAMLParser.CanParse]`, `[datapackage_yaml_parser.go:DataPackageYAMLParser.ParseManifest]` — parses `datapackage.yaml`/`.yml` including `Resource.Fields` (`schema.fields`, `concept`, `concept_ref`, `okf:concept`, `rdfType`, `unit`, `categories`, `value_mapping`).
  - **`OKFMarkdownParser` ([`okf_markdown_parser.go`](./okf_markdown_parser.go))**: `[okf_markdown_parser.go:NewOKFMarkdownParser]`, `[okf_markdown_parser.go:OKFMarkdownParser.Name]`, `[okf_markdown_parser.go:OKFMarkdownParser.CanParse]`, `[okf_markdown_parser.go:OKFMarkdownParser.ParseManifest]`, and `[okf_markdown_parser.go:ParseOKFFrontmatter]`.
  - **`LLMSTxtParser` ([`llmstxt_parser.go`](./llmstxt_parser.go))**: `[llmstxt_parser.go:NewLLMSTxtParser]`, `[llmstxt_parser.go:LLMSTxtParser.Name]`, `[llmstxt_parser.go:LLMSTxtParser.CanParse]`, `[llmstxt_parser.go:LLMSTxtParser.ParseManifest]`, and `[llmstxt_parser.go:ParseCatalog]` supporting flat `- [Title](URI): Desc` entries and hierarchical `### [REQ-XX]` requirement blocks.
  - **`OpenAPI & CLI Synthesizer` ([`openapi_and_cli_synthesizer.go`](./openapi_and_cli_synthesizer.go))**: `[openapi_and_cli_synthesizer.go:HydrateExternalCLIContract]`, `[openapi_and_cli_synthesizer.go:SynthesizeContractMarkdown]`, `[openapi_and_cli_synthesizer.go:StripOJSBlocks]` (plus internal `[openapi_and_cli_synthesizer.go:synthesizeOpenAPIMarkdown]`).

## Invariants

- **Specific Concept File Preservation (`[unified_reader.go:UnifiedMetadataReader.ReadNode]`)**: When `ReadNode` targets a specific `.md` file other than `index.md`, it preserves that file's title and body without overwriting it from sibling `datapackage.*` files, defaulting `manifest.Type` to `"concept"`.
- **Non-Eclipsing Contracts (`[unified_reader.go:UnifiedMetadataReader.ReadNode]`)**: External `owl-cli`/`okf-cli` and `openapi` contracts hydrate via `HydrateExternalCLIContract` and `SynthesizeContractMarkdown` without preventing subsequent `datapackage` contracts (`len(pkg.Resources) > 0`) from loading tabular resources.
- **Multi-Link `llms.txt` Regex (`[llmstxt_parser.go:ParseCatalog]`)**: Matches the first Markdown link `\[([^\]]+)\]\(([^)\s]+)\)` on a bullet line even when secondary ontology links follow on the same line.

## Side Effects

- **Filesystem & HTTP Reads (`[unified_reader.go:UnifiedMetadataReader.readData]`)**: Reads local files/directories (`os.Stat`, `os.ReadFile`) and issues outbound HTTP `GET` requests for `http(s)://` URIs.

## Verification

```bash
cd src && go test ./adapters/outbound/manifests/...
```
