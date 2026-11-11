# C-Shared FFI Adapter (`src/adapters/inbound/cshared`)

## Overview & Role

C-ABI shared library adapter (`package main`, `import "C"` in [`exports.go`](./exports.go)) exposing Open Knowledge Format (OKF) / DataMesh use cases and cryptographic/policy domain operations to host runtimes (such as Python `ctypes` in `datamesh-sdk`).

## Interface

### Lifecycle & Memory ([`exports.go`](./exports.go))
- **`[exports.go:DataMeshInit]`**: `DataMeshInit(configJSON *C.char) *C.char` — Thread-safe (`sync.Once`) runtime initialization overlaying optional JSON config (`catalog_url`, `storage_path`, `adapters`). Falls back to `os.TempDir()/datamesh-cache` if primary storage path fails.
- **`[exports.go:DataMeshFreeString]`**: `DataMeshFreeString(ptr *C.char)` — Frees C strings allocated by `makeJSONResponse` (`C.free(unsafe.Pointer(ptr))`).

### Catalog, Resolution & Querying ([`exports.go`](./exports.go))
- **`[exports.go:DataMeshDiscoverCatalog]`**: `DataMeshDiscoverCatalog(catalogURL *C.char) *C.char`
- **`[exports.go:DataMeshSearchCatalog]`**: `DataMeshSearchCatalog(keyword *C.char, catalogURL *C.char) *C.char`
- **`[exports.go:DataMeshResolveDataProduct]`**: `DataMeshResolveDataProduct(uri *C.char) *C.char`
- **`[exports.go:DataMeshQueryResource]`**: `DataMeshQueryResource(resourceURI *C.char, queryOptionsJSON *C.char) *C.char`
- **`[exports.go:DataMeshExecuteSQL]`**: `DataMeshExecuteSQL(sqlQuery *C.char, optionsJSON *C.char) *C.char`

### Validation, Lineage, Diffing & Attestation ([`exports.go`](./exports.go))
- **`[exports.go:DataMeshValidate]`**: `DataMeshValidate(target *C.char) *C.char`
- **`[exports.go:DataMeshGetLineage]`**: `DataMeshGetLineage(target *C.char, optionsJSON *C.char) *C.char`
- **`[exports.go:DataMeshLineageGraph]`**: `DataMeshLineageGraph(optionsJSON *C.char) *C.char`
- **`[exports.go:DataMeshDiffBundles]`**: `DataMeshDiffBundles(oldPath *C.char, newPath *C.char) *C.char`
- **`[exports.go:DataMeshAttestComputation]`**: `DataMeshAttestComputation(conceptPath *C.char, actualSQL *C.char, actualResultJSON *C.char) *C.char`

### Policy, Cryptography & Semantic Alignment ([`exports.go`](./exports.go))
- **`[exports.go:DataMeshEvaluatePolicy]`**: `DataMeshEvaluatePolicy(policyStr *C.char, sqlQueryStr *C.char, identityJSON *C.char) *C.char`
- **`[exports.go:DataMeshComputeEffectivePolicy]`**: `DataMeshComputeEffectivePolicy(policyJSON *C.char, identityJSON *C.char) *C.char`
- **`[exports.go:DataMeshVerifySignature]`**: `DataMeshVerifySignature(pubKeyHex *C.char, messageStr *C.char, signatureHex *C.char) *C.char`
- **`[exports.go:DataMeshParseOpenSSHKey]`**: `DataMeshParseOpenSSHKey(sshKeyStr *C.char) *C.char`
- **`[exports.go:DataMeshVerifyGitHubIdentity]`**: `DataMeshVerifyGitHubIdentity(didStr *C.char, claimedPubKey *C.char, githubKeysRaw *C.char) *C.char`
- **`[exports.go:DataMeshResolveConceptCorrespondences]`**: `DataMeshResolveConceptCorrespondences(sourceMappingJSON *C.char, targetMappingJSON *C.char, conceptCatalogJSON *C.char) *C.char`
- **`[exports.go:DataMeshGenerateCrossDatasetJoinSQL]`**: `DataMeshGenerateCrossDatasetJoinSQL(sourceMappingJSON *C.char, sourceAlias *C.char, targetMappingJSON *C.char, targetAlias *C.char, conceptCatalogJSON *C.char) *C.char`

## Invariants

- **Lazy Initialization (`[exports.go:ensureInit]`)**: Every exported function calls `ensureInit()`, guaranteeing `initRuntime("")` runs at least once via `sync.Once` even if `DataMeshInit` was not called explicitly.
- **Uniform JSON Envelope (`[exports.go:makeJSONResponse]`)**: All functions returning `*C.char` serialize `{"success": true, "data": ...}` or `{"success": false, "error": "..."}`.
- **Explicit C-Heap Ownership (`[exports.go:DataMeshFreeString]`)**: Every returned `*C.char` is allocated via `C.CString` and must be freed by the host language via `DataMeshFreeString`.

## Side Effects

- **Heap Allocation & Runtime State**: Allocates native C heap memory (`C.CString`/`C.free`), mutates package-level singleton use case state protected by `sync.Once` (`[exports.go:once]`), and delegates filesystem/network I/O to underlying use cases.

## Verification

```bash
cd src && go test ./adapters/inbound/cshared/...
```
