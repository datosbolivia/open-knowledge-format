# `okf` CLI Entrypoint (`src/cmd/okf`)

## Purpose & Architectural Role

Package `main` ([`main.go`](./main.go)) is the primary command-line entrypoint for the **Open Knowledge Format (OKF/ODKF v0.2) Engine & CLI** (`v0.3.0-dev.1`). It acts as an inbound CLI driving adapter that parses flexible POSIX/Go flags, dispatches subcommands to application use cases in [`src/usecases`](../../usecases/README.md), and formats human-readable or JSON reports to `stdout`/`stderr`.

## Interface

Binary invocation:

```bash
okf <command> [arguments]
```

| Subcommand | Handler (`[file:symbol]`) | Flags | Arguments | Delegated Use Case / Adapter | Description |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `init` | `[main.go:runInit]` | `-title`, `-desc` | `[target_dir]` (default `.`) | `usecases.InitBundle` | Initialize a new OKF v0.2 bundle directory structure (`index.md`, `log.md`, `concepts/sample_dataset.md`). |
| `add` | `[main.go:runAdd]` | `-type` (`concept`), `-title`, `-desc`, `-author`, `-cadence` (`P1M`) | `<concept_path>` | `usecases.AddConcept` | Create a new concept Markdown document with frontmatter template. |
| `link` | `[main.go:runLink]` | `-rel` | `<source.md> <target.md>` | `usecases.LinkConcepts` | Insert a semantic link between two concepts in the knowledge graph. |
| `run` | `[main.go:runRun]` | `-target`, `-json`, `-bundle` | `<script_path> [script_args...]` | `usecases.ExecuteScript` | Execute a pipeline script, compute SHA-256 provenance, and record execution in `log.md`. |
| `audit` | `[main.go:runAudit]` | `-json` | `[target_dir]` (default `.`) | `usecases.AuditBundle` | Audit bundle for broken links, orphan nodes, circular references, and concept staleness. |
| `graph` | `[main.go:runGraph]` | *(none)* | `[target_dir]` (default `.`) | `usecases.BuildGraphAdjacency` | Extract directed adjacency graph of concepts and semantic relations as indented JSON. |
| `index` | `[main.go:runIndex]` | `-json` | `[target_dir]` (default `.`) | `usecases.GenerateIndex` | Scan bundle and generate/update `index.md` for progressive disclosure. |
| `validate` | `[main.go:runValidate]` | `-json`, `-timeout` (`5`s) | `<target...>` (file, dir, or URL) | `usecases.NewValidateUseCase` | Validate one or more concept documents, bundle directories, or URLs against OKF v0.2 rules. |
| `diff` | `[main.go:runDiff]` | `-json` | `<bundle_v1> <bundle_v2>` | `usecases.DiffBundles` | Classify schema and concept changes between two bundle versions (`PATCH`/`MINOR`/`MAJOR`). |
| `attest` | `[main.go:runAttest]` | `-sql`, `-param` (`k=v,...`) | `<concept_file.md>` | `usecases.AttestComputation` | Verify an Attested Computation receipt or SQL query and emit a JSON `AttestationVerdict`. |
| `visualize` | `[main.go:runVisualize]` | `-bundle`, `-out`, `-name` | `[bundle_dir]` | `usecases.NewVisualizeUseCase` | Generate a self-contained interactive HTML graph visualization (defaults to `<bundle>/viz.html`). |
| `inspect` | `[main.go:runInspect]` | `-json` | `<target_path>` | `manifests.NewUnifiedMetadataReader`, `domain.ValidateConceptContent`, `usecases.NewValidateUseCase` | Display parsed frontmatter, SKOS ontology anchors, contracts, resources, and validation status. |
| `version` | `[main.go:main]` | `--version`, `-v` | *(none)* | Constant `[main.go:version]` (`0.3.0-dev.1`) | Print CLI version and specification engine identifier. |
| `help` | `[main.go:printUsage]` | `--help`, `-h` | *(none)* | Built-in usage printer | Print CLI usage summary to `stderr`. |

## Invariants

1. **Flexible Flag Ordering (`[main.go:parseFlexibleFlags]`)**: Flags and positional arguments may be interleaved in any order (e.g., `okf link source.md target.md -rel "Correlates With"`). Non-boolean flags consume the subsequent token if it does not start with `-`. Recognized boolean flags in `[main.go:isBooleanFlag]` are `json`, `strict`, `v`, `quiet`, `h`, `help`, or any `-flag=value` token.
2. **Exit Code Determinism**:
   - Missing required arguments, unknown commands, or failed use case invocations exit with status `1` (`os.Exit(1)`).
   - `okf audit` (`[main.go:runAudit]`) exits `1` when `!report.Passed` (including in `-json` mode).
   - `okf validate` (`[main.go:runValidate]`) iterates across all provided targets and exits `1` if any target errors or reports `!report.Valid`.
   - `okf run` (`[main.go:runRun]`) propagates the underlying script's non-zero exit code (`os.Exit(res.ExitCode)`).
   - `okf attest` (`[main.go:runAttest]`) always emits JSON to `stdout` and exits `1` on error or when `verdict.Verdict` is `"refuted"` or `"error"`.

## Side Effects

- **Process I/O (`[main.go:main]`)**: Reads CLI arguments from `os.Args`; writes structured JSON and standard reports to `os.Stdout`; writes usage text (`[main.go:printUsage]`), `visualize` summary metrics, and error diagnostics to `os.Stderr`; terminates the process via `os.Exit(1)` or `os.Exit(res.ExitCode)`.
- **Filesystem & Subprocess Delegation**: Mutates local bundle files via delegated use cases during `init`, `add`, `link`, `index`, `run`, and `visualize`; spawns external script subprocesses in `run`; performs local file reads or outbound HTTP requests (with configurable `-timeout`, default `5s`) during `validate` and `inspect`.

## Verification

```bash
cd src && go test ./...
cd src && go build ./cmd/okf
```
