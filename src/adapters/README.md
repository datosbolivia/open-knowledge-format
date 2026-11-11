# Hexagonal Adapters Layer (`src/adapters`)

## Overview & Role

`src/adapters` is an Intermediate Blueprint (Hay) containing primary (inbound) and secondary (outbound) adapters wiring the Open Knowledge Format (OKF) / DataMesh core domain ([`src/domain`](../domain/README.md)) and use cases ([`src/usecases`](../usecases/README.md)) to external interfaces, runtimes, file formats, and storage backends via [`src/ports`](../ports/README.md).

## Subtree Index

| Subtree | Role | Child Packages |
| :--- | :--- | :--- |
| [`inbound/`](./inbound/README.md) | Primary driving adapters (CLI, HTTP/MCP Server, C-Shared FFI, WebAssembly) | [`./inbound/cli/README.md`](./inbound/cli/README.md), [`./inbound/cshared/README.md`](./inbound/cshared/README.md), [`./inbound/wasm/README.md`](./inbound/wasm/README.md) |
| [`outbound/`](./outbound/README.md) | Secondary driven adapters (Config, Tabular Engine, Manifests, Resolvers, Storage) | [`./outbound/config/README.md`](./outbound/config/README.md), [`./outbound/engine/README.md`](./outbound/engine/README.md), [`./outbound/manifests/README.md`](./outbound/manifests/README.md), [`./outbound/resolvers/README.md`](./outbound/resolvers/README.md), [`./outbound/storage/README.md`](./outbound/storage/README.md) |

## Interface

- **Inbound Driving Surface ([`./inbound/README.md`](./inbound/README.md))**: `datamesh` CLI & HTTP/MCP server (`[main.go:main]`, `[serve.go:StartDataMeshServer]`), C-ABI exports (`[exports.go:DataMeshInit]`–`[exports.go:DataMeshFreeString]`), and WebAssembly `globalThis.DataMesh` (`[main_wasm.go:main]`).
- **Outbound Driven Surface ([`./outbound/README.md`](./outbound/README.md))**: `FileAndEnvConfigLoader`, `InMemTabularQueryEngine`, `QueryEngineRegistry`, `UnifiedMetadataReader`, `DataPackageJSONParser`, `DataPackageYAMLParser`, `OKFMarkdownParser`, `LLMSTxtParser`, `LLMSTxtCatalogResolver`, `NodeDataProductResolver`, `FileStorage`.

## Invariants

- **Hexagonal Dependency Rule**: Outbound adapters depend only on [`src/domain`](../domain/README.md) and [`src/ports/outbound`](../ports/outbound/README.md); inbound adapters orchestrate [`src/usecases`](../usecases/README.md) and [`src/ports/inbound`](../ports/inbound/README.md).
- **SSRF & Local File Guards**: Inbound HTTP/MCP endpoints ([`./inbound/cli/README.md`](./inbound/cli/README.md)) reject private/loopback/link-local/multicast IPs (`hostIsBlocked`, `validateOutboundURL`) and forbid local filesystem paths or traversal (`isForbiddenLocalRef`, `isRemoteOrTriad`).
- **Deterministic Fallbacks & Atomic Storage**: `InMemTabularQueryEngine` rejects unsupported SQL clauses and multi-table queries explicitly; `FileStorage` enforces base-directory confinement (`resolveSafePath`) and writes via `.tmp` + `os.Rename`.

## Side Effects

- Aggregates CLI I/O, TCP/HTTP listener binding, outbound HTTP requests, C-heap allocations, WASM global binding, and local filesystem storage across [`./inbound/README.md`](./inbound/README.md) and [`./outbound/README.md`](./outbound/README.md).

## Verification

```bash
cd src && go test ./adapters/...
```
