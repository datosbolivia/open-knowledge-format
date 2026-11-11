# Command-Line Entrypoints (`src/cmd`)

## Architectural Overview

The `src/cmd` directory is an Intermediate Blueprint (Hay) housing standalone executable entrypoints (`package main`) for the Open Knowledge Format (OKF/ODKF v0.2) Go module (`github.com/datosbolivia/open-knowledge-format/src`). Entrypoints in this subtree act strictly as thin CLI driving adapters that wire command-line flags and positional arguments into [`src/usecases`](../usecases/README.md), [`src/domain`](../domain/README.md), and [`src/adapters/outbound`](../adapters/outbound/README.md).

*(Note: The DataMesh CLI/MCP server entrypoint resides under [`src/adapters/inbound/cli`](../adapters/inbound/cli/README.md), whereas `src/cmd` provides the standalone `okf` specification engine binary.)*

## Subtree Index

| Directory | Binary / Package | Role & Contract Summary |
| :--- | :--- | :--- |
| [`okf/`](./okf/README.md) | `okf` (`package main`) | Standalone OKF/ODKF v0.2 Engine & CLI (`v0.2.2-rc.2`) supporting bundle initialization (`init`), authoring (`add`, `link`, `index`), governance (`validate`, `audit`, `diff`, `attest`), provenance tracking (`run`), and inspection (`graph`, `visualize`, `inspect`). |

## Interface

- **Child Package Contracts**:
  - [`./okf/README.md`](./okf/README.md): `[main.go:main]`, `[main.go:runInit]`, `[main.go:runAdd]`, `[main.go:runLink]`, `[main.go:runRun]`, `[main.go:runAudit]`, `[main.go:runGraph]`, `[main.go:runIndex]`, `[main.go:runValidate]`, `[main.go:runDiff]`, `[main.go:runAttest]`, `[main.go:runVisualize]`, `[main.go:runInspect]`, `[main.go:parseFlexibleFlags]`, `[main.go:isBooleanFlag]`.

## Invariants

- **Thin Driving Layer**: Packages under `src/cmd` do not implement domain validation rules or storage parsing directly; business logic is delegated to [`src/usecases`](../usecases/README.md) and [`src/domain`](../domain/README.md).
- **Deterministic CLI Contracts**: Every subcommand in [`./okf/README.md`](./okf/README.md) supports position-independent flag parsing (`[main.go:parseFlexibleFlags]`), structured `-json` output where applicable, and non-zero exit codes (`os.Exit`) on validation, audit, attestation, or script execution failures.

## Side Effects

- Delegates process arguments (`os.Args`), standard streams (`os.Stdout`, `os.Stderr`), filesystem mutations, and HTTP/subprocess execution through child entrypoint [`./okf/README.md`](./okf/README.md).

## Verification

```bash
cd src && go test ./...
cd src && go build ./cmd/...
```
