# Architectural Ports (`src/ports`)

## Overview & Role

`src/ports` is an Intermediate Blueprint (Hay) containing the hexagonal boundary interfaces connecting the pure domain model ([`src/domain`](../domain/README.md)) and application orchestration layer ([`src/usecases`](../usecases/README.md)) to external driving and driven adapters ([`src/adapters`](../adapters/README.md)). In accordance with [`SPEC.md`](../../SPEC.md) §16.1 (*Hexagonal Port Separation Principle*), all contracts in this subtree are technology-agnostic Go interfaces that invert dependencies between core domain logic and infrastructure.

## Subtree Index

| Directory | Package | Role & Scope | Contract Document |
| :--- | :--- | :--- | :--- |
| [`inbound/`](./inbound/README.md) | `inbound` | Primary (driving) service ports consumed by CLI, C-Shared FFI, and WASM adapters (`CatalogServicePort`, `DataProductServicePort`, `QueryServicePort`, `ConfigServicePort`, `PublicationServicePort`, `SemanticServicePort`, `DiscoveryServicePort`, `LineageServicePort`, `InteroperabilityServicePort`). | [`./inbound/README.md`](./inbound/README.md) |
| [`outbound/`](./outbound/README.md) | `outbound` | Secondary (driven) infrastructure ports implemented by outbound adapters for config, storage, manifest parsing, catalog/node/triad resolution, dataset publication, and SQL/tabular query engines. | [`./outbound/README.md`](./outbound/README.md) |

## Interface

- **Inbound Service Ports ([`./inbound/README.md`](./inbound/README.md))**: `[services.go:CatalogServicePort]`, `[services.go:DataProductServicePort]`, `[services.go:QueryServicePort]`, `[services.go:ConfigServicePort]`, `[services.go:PublicationServicePort]`, `[services.go:SemanticServicePort]`, `[services.go:DiscoveryServicePort]`, `[services.go:LineageServicePort]`, `[services.go:InteroperabilityServicePort]`.
- **Outbound Infrastructure Ports ([`./outbound/README.md`](./outbound/README.md))**: `[config.go:ConfigLoaderPort]`, `[harvester.go:AuthProviderPort]`, `[harvester.go:WellKnownResolverPort]`, `[package_reader.go:ManifestParserPort]`, `[package_reader.go:UnifiedMetadataReaderPort]`, `[publisher.go:PublisherTargetPort]`, `[query.go:QueryEnginePort]`, `[query.go:QueryEngineRegistryPort]`, `[resolvers.go:CatalogResolverPort]`, `[resolvers.go:DataProductResolverPort]`, `[triad_resolver.go:TriadResolverPort]`, `[storage.go:StoragePort]`.

## Invariants

1. **Zero Infrastructure Coupling**: Packages under `src/ports` import only `context` and [`github.com/datosbolivia/open-knowledge-format/src/domain`](../domain/README.md). They never import `src/usecases` or `src/adapters`.
2. **Pure Contract Layer**: No structs, state, or executable logic reside in `src/ports`; only `interface` definitions are permitted.
3. **Normative Spec Alignment**: Implements the abstract port separation specified in ODKF v0.2 ([`SPEC.md`](../../SPEC.md) §16.1–16.2), enabling plug-and-play adapters across CLI, embedded library, and WebAssembly targets.

## Side Effects

- None (`["None"]` — pure interface declarations).

## Verification

```bash
cd src && go test ./...
```
