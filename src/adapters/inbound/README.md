# Inbound Adapters (`src/adapters/inbound`)

## Overview & Role

`src/adapters/inbound` is an Intermediate Blueprint (Hay) containing the primary driving adapters translating CLI invocations, HTTP REST requests, Model Context Protocol (MCP) JSON-RPC 2.0 calls, C-ABI FFI calls, and browser/Node WebAssembly calls into [`src/domain`](../../domain/README.md) and [`src/usecases`](../../usecases/README.md) executions.

## Subtree Index

| Subpackage | Target Artifact | Public Surface |
| :--- | :--- | :--- |
| [`cli/`](./cli/README.md) | `datamesh` binary & HTTP/MCP server | Commands: `catalog`, `search`, `get`, `validate`, `sql`, `query`, `lineage`, `serve`, `config`. HTTP routes: `/health`, `/.well-known/datamesh.json`, `/.well-known/okf.json`, `/proxy`, `/api/catalog`, `/api/sql`, `/mcp`. |
| [`cshared/`](./cshared/README.md) | `c-shared` dynamic library | Exported C symbols: `DataMeshInit`, `DataMeshDiscoverCatalog`, `DataMeshSearchCatalog`, `DataMeshResolveDataProduct`, `DataMeshQueryResource`, `DataMeshExecuteSQL`, `DataMeshValidate`, `DataMeshGetLineage`, `DataMeshLineageGraph`, `DataMeshEvaluatePolicy`, `DataMeshVerifySignature`, `DataMeshComputeEffectivePolicy`, `DataMeshParseOpenSSHKey`, `DataMeshVerifyGitHubIdentity`, `DataMeshResolveConceptCorrespondences`, `DataMeshGenerateCrossDatasetJoinSQL`, `DataMeshDiffBundles`, `DataMeshAttestComputation`, `DataMeshFreeString`. |
| [`wasm/`](./wasm/README.md) | `js/wasm` module | Global `DataMesh` JS object (`discover`, `search`, `resolve`, `get`, `query`, `executeSQL`, `validate`, `getConfig`). |

## Interface

- **CLI & HTTP/MCP Adapter ([`./cli/README.md`](./cli/README.md))**: `[main.go:main]`, `[serve.go:ServerConfig]`, `[serve.go:StartDataMeshServer]`, `[security.go:hostIsBlocked]`, `[security.go:validateOutboundURL]`, `[security.go:isRemoteOrTriad]`, `[security.go:isForbiddenLocalRef]`.
- **C-Shared FFI Adapter ([`./cshared/README.md`](./cshared/README.md))**: `[exports.go:DataMeshInit]` through `[exports.go:DataMeshFreeString]`.
- **WebAssembly Adapter ([`./wasm/README.md`](./wasm/README.md))**: `[main_wasm.go:main]`, `globalThis.DataMesh`.

## Invariants

- **Outbound Proxy & MCP SSRF Protection ([`./cli/README.md`](./cli/README.md))**: `validateOutboundURL` permits only `http`/`https` schemes with non-empty hostnames that do not resolve to loopback, private, link-local, multicast, or unspecified IPs (`hostIsBlocked`).
- **MCP Resource Isolation ([`./cli/README.md`](./cli/README.md))**: `datamesh_query_resource` enforces `isRemoteOrTriad`, blocking local paths (`/etc/passwd`, `file://`, Windows drive letters, UNC paths, `..` segments).
- **FFI Memory Ownership ([`./cshared/README.md`](./cshared/README.md))**: Every `*C.char` returned by `DataMesh*` functions is heap-allocated via `C.CString` with envelope `{"success": bool, "data": ..., "error": ...}` and must be released by the caller via `DataMeshFreeString`.

## Side Effects

- Aggregates process, network, C-heap, and JavaScript global state side effects across child packages [`./cli/README.md`](./cli/README.md), [`./cshared/README.md`](./cshared/README.md), and [`./wasm/README.md`](./wasm/README.md).

## Verification

```bash
cd src && go test ./adapters/inbound/...
```
