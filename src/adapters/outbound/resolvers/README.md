# Catalog & Node Resolvers (`src/adapters/outbound/resolvers`)

## Overview & Role

Implements [`outbound.CatalogResolverPort`](../../../ports/outbound/README.md) (`LLMSTxtCatalogResolver` in [`llmstxt_resolver.go`](./llmstxt_resolver.go)) and [`outbound.DataProductResolverPort`](../../../ports/outbound/README.md) (`NodeDataProductResolver` in [`node_resolver.go`](./node_resolver.go)) for discovering federated `llms.txt` catalogs and resolving OKF v0.2 Data Products.

## Interface

- **`LLMSTxtCatalogResolver` ([`llmstxt_resolver.go`](./llmstxt_resolver.go))**:
  - `[llmstxt_resolver.go:NewLLMSTxtCatalogResolver]`: `NewLLMSTxtCatalogResolver(timeout time.Duration) *LLMSTxtCatalogResolver` (defaults to `30s` if `<= 0`).
  - `[llmstxt_resolver.go:LLMSTxtCatalogResolver.FetchCatalog]`: `(r *LLMSTxtCatalogResolver) FetchCatalog(ctx context.Context, catalogURL string) (*domain.Catalog, error)` — Fetches `http(s)://` or `file://` `llms.txt` content and parses flat and hierarchical catalog entries, resolving relative URIs against `catalogURL`.
- **`NodeDataProductResolver` ([`node_resolver.go`](./node_resolver.go))**:
  - `[node_resolver.go:NewNodeDataProductResolver]`: `NewNodeDataProductResolver(reader outbound.UnifiedMetadataReaderPort) *NodeDataProductResolver`
  - `[node_resolver.go:NodeDataProductResolver.FetchDataProduct]`: `(r *NodeDataProductResolver) FetchDataProduct(ctx context.Context, resolvedURL string) (*domain.DataProduct, error)` — Delegates to `reader.ReadNode(ctx, resolvedURL)`.

## Invariants

- **Port Decoupling (`[node_resolver.go:NodeDataProductResolver]`)**: Depends strictly on `outbound.UnifiedMetadataReaderPort` rather than concrete `manifests` implementations.
- **Multi-Link `llms.txt` Resolution (`[llmstxt_resolver.go:parseLLMSTxt]`)**: Extracts the primary `[Title](URI)` link using non-greedy URL matching `([^)\s]+)` and resolves relative paths against the catalog's base URL.

## Side Effects

- **Network & Filesystem I/O (`[llmstxt_resolver.go:LLMSTxtCatalogResolver.FetchCatalog]`)**: Issues HTTP `GET` requests (`net/http`) or reads `file://` paths (`os.ReadFile`).

## Verification

```bash
cd src && go test ./adapters/outbound/resolvers/...
```
