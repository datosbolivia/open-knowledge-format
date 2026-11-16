# Open Knowledge Format (OKF) v0.3 Specification Hardening & Codebase Alignment

*   **Author:** andreschirinos
*   **Status:** Approved
*   **Date:** 2026-11-15

--------------------------------------------------------------------------------

## Overview

Open Knowledge Format ([`SPEC.md`](file:///home/andreschirinos/Proyectos/open-knowledge-format/SPEC.md)) defines a markdown-and-YAML convention for packaging documentation, dataset contracts, provenance, trust metadata, and reproducible SQL computations, enabling serverless inspection and querying of data bundles by human operators and AI agents. This design rectifies structural, normative, security, and implementation drift defects across [`SPEC.md`](file:///home/andreschirinos/Proyectos/open-knowledge-format/SPEC.md) (currently labeled v0.2 while implementations run v0.3.0-dev.1), the Go reference engine (`open-knowledge-format/src/`), and the sibling Python SDK (`../datamesh-sdk/bindings/python/`).

## Problem Statement

Although commit `47ce96f` extracted initial P2P and state-store documents into [`docs/rfcs/0001-p2p-sovereign-network.md`](file:///home/andreschirinos/Proyectos/open-knowledge-format/docs/rfcs/0001-p2p-sovereign-network.md) and [`docs/architecture/operational-state-store.md`](file:///home/andreschirinos/Proyectos/open-knowledge-format/docs/architecture/operational-state-store.md), an audit of [`SPEC.md`](file:///home/andreschirinos/Proyectos/open-knowledge-format/SPEC.md) (1,513 lines) and reference implementations exposed seven blocker defects and sixteen major inconsistencies causing parser errors, validation contradictions, and cross-runtime drift:

1.  **Broken Document Structure in `SPEC.md`**: Four outer code blocks in §10.2 and Appendix A ([`SPEC.md`](file:///home/andreschirinos/Proyectos/open-knowledge-format/SPEC.md) lines 655–690, 1344–1377, 1425–1471, 1476–1513) use triple-backtick fences around inner triple-backtick SQL blocks, prematurely closing the examples, swallowing section headings (`### 10.3`, `### v0.2 Form`), and leaving an unclosed fence at EOF.
2.  **Self-Contradictory Normative Validation Rules** (across Producer Enforcement rules `PE1`–`PE5`, Consumer/Operational Enforcement invariants `OE1`–`OE5`, and validation checks `RULE-01`–`RULE-12` in Open Data Knowledge Format, `ODKF`):
    *   `RULE-08` (§14.5, line 1099) enforces a colon-only regular expression (`^[a-zA-Z0-9_.-]+:[a-zA-Z0-9_. -]+...`) that rejects four of the five canonical triad representations mandated in §14.4 and `OE4` (slash paths, reverse-DNS `@`, `okf://` URIs, and direct HTTP URLs).
    *   §11 (line 797) permits a bundle-root `index.md` to declare `type: dcat:Dataset` as a Root Concept Manifest, conflicting with §3.1, §8, and `PE3` (§14.5, line 1074), which forbid concept frontmatter inside `index.md` without stating the namespace-based rule implemented in [`src/domain/validator.go`](file:///home/andreschirinos/Proyectos/open-knowledge-format/src/domain/validator.go#L537-L560).
    *   `RULE-09` (§14.5, line 1100) requires `agent:<name>`, `<email>`, or a URI for actors, rejecting the `<producer>/<version>` and `process:<id>` actor syntax defined in §7 and used across specification examples.
    *   §14.2 (lines 980–981) places a remote `https://` URI inside `contracts[].path`, violating `PE4` and `RULE-04` (`path` for local bundle files vs. `ref` for remote HTTP/S URIs).
3.  **Security & Integrity Gaps**:
    *   **Dangerous URL Schemes**: Remote URL resolution across `sources`, `contracts`, and `/.well-known/` discovery lacks a normative scheme guard rejecting `file://`, `gopher://`, or `ftp://` URIs.
    *   **Attester Sandboxing Contradiction**: §10.2 mandates WASI/container sandboxing for `attester` scripts while §12.3 simultaneously defers the sandboxing ABI, even though the reference `okf attest` implementation ([`src/usecases/diff_and_attest.go`](file:///home/andreschirinos/Proyectos/open-knowledge-format/src/usecases/diff_and_attest.go#L317-L397)) performs static SQL AST normalization without executing scripts.
    *   **Path Traversal & Zip Slip**: Local contract paths (§6.2) and archive `#fragment` extraction (§14.4.1) lack normative bundle-root containment and `..` segment rejection inside archive entries.
    *   **Non-Deterministic YAML Signatures & Microdata Leakage**: Embedding `signatures[]` with `verified: true` inside YAML frontmatter lacks a cross-language canonical serialization rule (such as RFC 8785 JCS); `OE3` (`zero_microdata`) bans `SELECT *` without mandating `HAVING COUNT(*) >= min_group_size` in `SPEC.md`.
4.  **Incomplete Layering & Schema Gaps**:
    *   Part I (Core OKF, §1–§13) claims to be tool-free and self-contained, yet §11 and §12 embed CLI wire contracts (`okf attest`, `okf diff`) and forward-reference Part II concepts (`catalog:dataset:resource`, `RULE-01`–`RULE-12`, Kademlia DHT).
    *   Normative fields referenced in prose (`unit`, `aggregation_behavior`, `min_group_size`, `rdf_type`, `semantic_fields`, `alt_labels`, `exact_match`, `close_match`) have no formal YAML schema definitions in §4.1 or §14.2.
    *   `lineage` uses four incompatible key shapes across §14.2 (`source`), `RULE-12` (`derived_from`), §17.3 (`sources`, `was_derived_from`), and §18.2 (`lineage_sources`).
    *   §17.2 and §18.2 reference [`docs/architecture/operational-state-store.md`](file:///home/andreschirinos/Proyectos/open-knowledge-format/docs/architecture/operational-state-store.md) (labeled *Informational*) while leaking `datamesh-sdk` vendor identifiers (`datamesh://`, `@dm.track_lineage`, `.datamesh/`, `/.well-known/datamesh.json`).
5.  **Implementation Drift Between `SPEC.md`, Go Core, and Python SDK**:
    *   **`skos:closeMatch` Join Violation**: §14.3.1 prohibits `skos:closeMatch` in automated SQL join generation to avoid Cartesian explosions, but both Go ([`src/usecases/publish_and_semantic_usecase.go`](file:///home/andreschirinos/Proyectos/open-knowledge-format/src/usecases/publish_and_semantic_usecase.go#L398-L415)) and Python ([`semantic_alignment.py`](file:///home/andreschirinos/Proyectos/datamesh-sdk/bindings/python/datamesh/usecases/semantic_alignment.py#L231-L242)) emit `CASE` join clauses over `closeMatch` correspondences.
    *   **Undocumented Wire Protocols**: The `OKF_PROVENANCE_OUT` JSONL IPC protocol ([`src/usecases/run.go`](file:///home/andreschirinos/Proyectos/open-knowledge-format/src/usecases/run.go#L253-L299)) and RFC 8615 `/.well-known/okf.json` discovery ([`src/domain/wellknown.go`](file:///home/andreschirinos/Proyectos/open-knowledge-format/src/domain/wellknown.go)) are implemented in code but absent from `SPEC.md`.
    *   **Parser & Validator Discrepancies**: Go and Python `llms.txt` parsers only match Spanish keywords (`Dominio:`, `Recursos:`) and ignore the English keywords (`Domain:`, `Resources:`) required by §14.6.3; reverse-DNS dot disambiguation (`"censo.2024".votos@bolivia`) fails due to unquoted first-dot splitting and unpaired outer quote trimming; [`diff_and_attest.go`](file:///home/andreschirinos/Proyectos/open-knowledge-format/src/usecases/diff_and_attest.go#L411-L418) rejects valid `..` relative paths permitted by §6.2; and `SPEC.md` targets version `0.2` while code runs `0.3.0-dev.1`.

## Goals

*   **Fix Identified Structural and Markdown Parsing Defects in `SPEC.md`**: Convert four outer nested code blocks to 4-backtick fences, fix skipped heading levels (`H1 -> H3`), standardize RFC 2119 uppercase keywords in `RULE-09`–`RULE-12`, and repair broken `§` cross-references.
*   **Eliminate Normative Contradictions**: Reconcile `RULE-08` with all five triad forms, state the context-free `dcat:*` namespace rule for `index.md` across §3.1/§8/§11/`PE3`, align `RULE-09` actor patterns with §7 (`<producer>/<version>`, `human:<id>`, `process:<id>`), unify `lineage` and `policy`/`policies` schemas, and define YAML schemas for `semantic_fields`, `unit`, `aggregation_behavior`, `min_group_size`, and `rdf_type`.
*   **Harden Security & Privacy Invariants**: Specify a normative URL Scheme Guard (`OE6`), enforce bundle-root path containment and Zip Slip rejection (`OE7`), specify RFC 8785 JCS frontmatter + body SHA-256 signature verification (removing producer-supplied `verified: true`), mandate `HAVING COUNT(*) >= min_group_size` in `OE3`, and align §10.2/§12.3/§17.1 around static SQL AST attestation (`okf attest`) and user-invoked local execution (`okf run`).
*   **Clean Up Layering & Vendor Leakage**: Separate Part I (tool-free format) from Part II (runtime, CLI wire formats, and abstract operational state entities), replace `datamesh` vendor tokens with `okf` equivalents (keeping legacy aliases only as backward-compatible notes), and document `OKF_PROVENANCE_OUT` JSONL and `/.well-known/okf.json`.
*   **Align Go and Python Reference Implementations with `SPEC.md`**: Filter out `skos:closeMatch` in automated SQL join generation (failing fast when zero exact correspondences remain), support bilingual `llms.txt` metadata and description cleaning (`Domain`/`Dominio`, `Resources`/`Recursos`), support quoted dataset identifiers in reverse-DNS triads, and allow valid bundle-internal `..` relative paths in `diff_and_attest.go`.

## Non-Goals

*   Implementing a Kademlia DHT / libp2p networking daemon or WASI container runtime in the Go or Python core engines.
*   Building a general-purpose RDF/OWL reasoner or guaranteeing lossless roundtrip conversion across arbitrary external metadata standards.
*   Replacing SQLite with an external client-server RDBMS for local operational state in `datamesh-sdk`.

## Resolved Design Decisions

*   **`index.md` Root Concept Manifest (`src/domain/validator.go` lines 537–560)**: Classification is context-free and driven by the `type` namespace prefix: a bundle-root `index.md` declaring an explicit DCAT class (`dcat:Dataset`, `dcat:DataService`, or `dcat:CatalogRecord`) is validated as a Root Concept Manifest, whereas shorthand `type: dataset` or `type: service` inside `index.md` or `log.md` emits `PE3-RESERVED-FILE` (`WARNING`).
*   **Version Baseline (`0.3`)**: Bump `SPEC.md` format baseline from `0.2` to `0.3` (`okf_version: "0.3"`) to match the `0.3.0-dev.1` reference implementations and §13.1 breaking changes.

## Critical User Journeys (CUJs)

### CUJ 1: Deterministic Bundle Validation Across All Triad & Actor Forms

*   **Given**: An ODKF bundle containing concepts with `<producer>/<version>` actors, relative `../computations/` links, and references using all five canonical triad notations (`:`, `/`, `@`, `okf://`, `https://`).
*   **When**: A user or CI pipeline runs `okf validate` against the bundle.
*   **Then**: Validation succeeds without false-positive errors on `RULE-08` or `RULE-09`, verifies that all relative paths stay within the bundle root (`OE7`), and emits a `PE4-REMOTE-IN-PATH` warning if an `http(s)://` URL is placed in `contracts[].path` instead of `contracts[].ref`.

### CUJ 2: Safe Cross-Dataset SQL Join Generation Without `closeMatch` Explosion

*   **Given**: Two datasets whose `semantic_fields` map categorical values to concepts linked via both `skos:exactMatch` and `skos:closeMatch`.
*   **When**: An agent or user invokes `GenerateCrossDatasetJoinSQL` (Go) or `generate_cross_dataset_join_sql` (Python).
*   **Then**: The generated SQL `CASE` join expression includes only verified `direct_concept_ref` (direct concept ID reference), `skos:exactMatch`, and `skos_anchor` (shared `skos` URI) correspondences, excluding `skos:closeMatch` pairs as required by §14.3.1, and returns an error (`ErrNoExactCorrespondences` / `ValueError`) rather than falling back to raw column equality if zero exact correspondences exist.

### CUJ 3: Bilingual `llms.txt` Discovery and Quoted Reverse-DNS Triad Resolution

*   **Given**: A federated catalog publishing English `(Domain: Finance. Resources: res_a, res_b)` entries in `/llms.txt` and a reverse-DNS query referencing `"censo.2024".votos@bolivia#seccion.csv`.
*   **When**: The Go or Python runtime parses `/llms.txt` and resolves the triad reference.
*   **Then**: Both English and Spanish metadata attributes are extracted and stripped from the entry description, and the triad parser resolves unquoted dataset `censo.2024`, resource `votos`, catalog `bolivia`, and archive fragment `seccion.csv`.

## Technical Plan (Architecture)

The design resolves the audit findings across four decoupled layers: (1) Document Structure & Layering in `SPEC.md`, (2) Normative Grammar & Schema Unification, (3) Security & Privacy Invariants, and (4) Reference Engine Alignment (Go Core and Python SDK).

```text
+-------------------------------------------------------------------------+
|                   Part I: Core OKF v0.3 (§1 - §13)                      |
|  Tool-free Markdown + YAML Frontmatter | Provenance, Trust, Lifecycle   |
|  Declarative Attested Computation Schema (§10) | Level 0 Core (§11)     |
+------------------------------------+------------------------------------+
                                     |
                                     v
+-------------------------------------------------------------------------+
|               Part II: ODKF v0.3 Profile (§14 - §18)                    |
|  5-Form Canonical Triads + Quoted Reverse-DNS + #fragment (§14.4)       |
|  Producer Invariants (PE1-PE5), Consumer Invariants (OE1-OE7),          |
|  Validation Rules (RULE-01..RULE-12)                                    |
|  CLI Wire Contracts (§14.7): okf attest, okf diff, OKF_PROVENANCE_OUT   |
|  Discovery: /llms.txt (EN/ES), /.well-known/okf.json, Abstract State    |
+------------------------------------+------------------------------------+
                                     |
                   +-----------------+-----------------+
                   |                                   |
                   v                                   v
+----------------------------------+   +----------------------------------+
|   Go Core (open-knowledge-format)|   | Python SDK (datamesh-sdk)        |
| - validator.go (PE/RULE/OE)      |   | - semantic_alignment.py          |
| - triad.go (quoted dot split)    |   |   (exclude skos:closeMatch SQL)  |
| - publish_and_semantic_usecase.go|   | - models.py (quoted dot split)   |
|   (exclude skos:closeMatch SQL)  |   | - llmstxt_catalog.py (EN/ES)     |
| - llmstxt_parser.go (EN/ES)      |   | - sqlite_lineage.py ($OKF_HOME,  |
| - diff_and_attest.go (bundle ..) |   |   WAL, busy_timeout, user_version|
+----------------------------------+   +----------------------------------+
```

### 1. Structural Repair & Part I / Part II Layering ([`SPEC.md`](file:///home/andreschirinos/Proyectos/open-knowledge-format/SPEC.md))

*   **Nested Code Fence Repair**: Upgrade the four outer markdown example blocks in §10.2 (lines 655–690) and Appendix A (lines 1344–1377, 1425–1471, 1476–1513) from triple-backtick to quadruple-backtick fences so inner SQL blocks no longer terminate the outer fence prematurely.
*   **Heading Hierarchy & Cross-References**: Promote `### Conformance & Keyword Definitions` (line 19) to `## Conformance & Keyword Definitions`; promote Appendix A subheadings `### v0.1 Form` and `### v0.2 Form` to `## A.1 v0.1 Form` and `## A.2 v0.2 Form`; fix section citations in §7 (`§14.5` -> `§14.2`), `RULE-07` (`§11` -> `§14.5`), `RULE-10` (`§5.4`), `RULE-11` (`§5.2, §5.5`), and `RULE-12` (`§5.4, §14.2`).
*   **Part I vs. Part II Decoupling**: Keep Part I (§1–§13) self-contained as the Level 0 Core document standard. Move CLI wire contracts (`okf attest` `AttestationVerdict` and `okf diff`) and Level 1/Level 2 profile requirements out of Part I into Part II (§14.5 and §14.7).
*   **Version Baseline Alignment**: Update `SPEC.md` header and §12 from `0.2` to `0.3` (`okf_version: "0.3"`), clarifying that `okf_version` uses `<major>.<minor>` format revisions while dataset `version` and `okf diff` follow SemVer 2.0.0 (`MAJOR.MINOR.PATCH`).
*   **Context-Free `index.md` Rule (`PE3` / §3.1 / §8 / §11)**: Document the deterministic rule implemented in [`src/domain/validator.go`](file:///home/andreschirinos/Proyectos/open-knowledge-format/src/domain/validator.go#L537-L560): `index.md` and `log.md` are reserved navigation/chronology files (`PE3-RESERVED-FILE` emits `WARNING` if shorthand `type: dataset` or `type: service` is declared); a bundle-root `index.md` MAY act as a Root Concept Manifest when it declares an explicit DCAT class (`type: dcat:Dataset | dcat:DataService | dcat:CatalogRecord`).

### 2. Normative Grammar, Validation Rules, and Schema Unification ([`SPEC.md`](file:///home/andreschirinos/Proyectos/open-knowledge-format/SPEC.md) §7, §14, §17, §18)

*   **`RULE-08` Five-Form Triad Grammar**: Replace the colon-only regex in §14.5 with the five-form grammar (colon, slash, reverse-DNS with optional double-quoted dataset segment for dot disambiguation, `(okf|odkf)://` URI with `datamesh://` as deprecated alias, and `http(s)://` URLs, all supporting optional `#fragment`).
*   **`RULE-09`–`RULE-12` Alignment**: Align `RULE-09` with §7 actor patterns (`<producer>/<version>`, `(agent|human|process|user|team):<id>`, `Name <email>`, `http(s)://` or `did:` URI); capitalize RFC 2119 keywords; split `RULE-11` into ISO 8601 timestamps (`generated.at`, `verified[].at`, `stale_after`, `freshness.expires`) vs. ISO 8601 durations (`freshness.cadence`); and specify `RULE-12` cycle severities (`ERROR` on `supersedes`/`superseded_by`/`relations` DAG cycles, `WARNING` on body markdown links).
*   **Unified Frontmatter Schemas (§14.2 & §14.3)**:
    *   Fix §14.2 contract example (`path: https://...` -> `ref: https://...`).
    *   Standardize `lineage` on `sources` (accepting singular `source` as deprecated alias), `transformations: [{step, tool, hash}]`, `version`, and `updated_at`, plus top-level `relations: [{rel: derived_from | supersedes | in_catalog, target}]`.
    *   Standardize `policy` / `policies`: if both are declared, `policy` is merged as the first element of `policies[]` and evaluated under most-restrictive-wins (`deny > zero_microdata > allow_all`).
    *   Define formal YAML schemas for `semantic_fields` (`field_name`, `concept_id`, `rdf_type`, `unit`, `aggregation_behavior: sum | avg | count | min | max | none`, `value_labels: [{raw_value, concept_id, alt_labels}]`) and SKOS correspondence types (`direct_concept_ref`, `skos:exactMatch`, `skos_anchor`, `skos:closeMatch`).
*   **Wire Protocols & Operational State (§14.6, §14.7, §17, §18)**:
    *   Document RFC 8615 catalog discovery at `/.well-known/okf.json` (with `/.well-known/datamesh.json` as fallback) in §14.6.
    *   Document `AttestationVerdict` and `okf diff` JSON schemas and exit codes in §14.7, and the `OKF_PROVENANCE_OUT` JSONL IPC contract in §17.1.
    *   Keep §18 focused on abstract operational state entities and designate [`docs/architecture/operational-state-store.md`](file:///home/andreschirinos/Proyectos/open-knowledge-format/docs/architecture/operational-state-store.md) as the normative SQLite DDL reference without duplicating SQL DDL inside `SPEC.md`.

### 3. Security & Privacy Invariants ([`SPEC.md`](file:///home/andreschirinos/Proyectos/open-knowledge-format/SPEC.md) §6.2, §7, §10.2, §14.4.1, §14.5, §17.1)

*   **`OE6` (URL Scheme Guard)**: Conforming runtimes dereferencing remote URLs (`sources`, `contracts[].ref`, `/.well-known/`, HTTP range reads) MUST permit only `https://` (and `http://` when configured for local/intranet catalogs), rejecting `file://`, `gopher://`, and `ftp://` schemes in remote reference fields. Private-IP and cloud-metadata (`169.254.169.254`) blocking is documented as an operational recommendation for multi-tenant hosted proxy deployments.
*   **`OE7` (Path Containment & Zip Slip Prevention)**: All local paths (`§6.2`, `contracts[].path`, `policies[].path`) MUST be lexically cleaned (`filepath.Clean`) relative to the declaring file's directory, resolved through symlinks when the target file exists on disk, and verified to remain strictly inside the bundle root directory. Archive `#fragment` extractors (`§14.4.1`) MUST reject any `#fragment` or archive member entry containing `..` segments or leading `/`.
*   **Static Attestation & Local Runner Boundary (`§10.2`, `§12.3`, `§17.1`)**: Normatively define `okf attest` as Static / Declarative Attestation (SQL AST normalization and SHA-256 comparison against `canonical_query_hash` with zero host script execution). Maintain the §12.3 deferral for sandboxed executable attester ABIs, and classify `okf run` (§17.1) as explicit user-invoked local process execution.
*   **Deterministic Signature Verification (`§7`, `§14.2`)**: Remove producer `verified: true` from frontmatter schema. Specify that in-band `signatures[].signature` is computed over the SHA-256 digest of the RFC 8785 JSON Canonicalization Scheme (JCS) representation of the parsed YAML frontmatter (excluding the `signatures` key) concatenated with `\n` and the raw UTF-8 bytes of the Markdown body.
*   **`OE3` (`zero_microdata` Cohort Threshold)**: Mandate that SQL engines enforcing `policy: zero_microdata` reject unaggregated row projections and enforce `HAVING COUNT(*) >= min_group_size` (default `5`) on grouped queries.

### 4. Reference Codebase Alignment (`open-knowledge-format` & `datamesh-sdk`)

*   **Exclude `skos:closeMatch` & Fail Fast**: Filter out `skos:closeMatch` correspondences during automated SQL join generation in Go and Python, and return an explicit error when zero verified/`exactMatch`/`skos_anchor` correspondences remain instead of falling back to raw column equality.
*   **Quoted Reverse-DNS Triad Normalization**: Preserve paired outer quotes, split `"dataset.with.dots".resource@catalog` after the closing quote, unquote the parsed `Dataset` segment, and quote dot-containing datasets in `ResourceTriad.SovereignDNS()`.
*   **Bilingual `llms.txt` Parsing & Truncation**: Match both English and Spanish labels (`Domain`/`Dominio`, `Resources`/`Recursos`) in extraction regexes, stop-token lookaheads, and description truncation.
*   **Bundle-Root Path Containment**: Resolve `bundleRoot` by walking upward from the concept file for `index.md` or `.git` (falling back to the concept file's directory) in both `validator.go` and `diff_and_attest.go`.
*   **Python State Store Alignment**: Resolve `OKF_METADATA_DB` -> `DATAMESH_METADATA_DB` -> `$OKF_HOME/metadata.db` -> `$DATAMESH_HOME/metadata.db` -> `.okf/metadata.db` -> `.datamesh/metadata.db` -> `~/.okf/metadata.db` -> `~/datamesh/metadata.db`, set `busy_timeout = 5000` per connection, and set `journal_mode = WAL` + `user_version = 1` during database initialization.

## Alternatives Considered

### Alternative 1: Retain Kademlia DHT & P2P Wire Protocol Inside Core `SPEC.md` (§15)

*   **Pros**: Keeps all sovereign networking concepts in a single file.
*   **Cons**: Adds ~400 lines of experimental libp2p/DHT protocol details with no production implementation in `src/`, and introduces an unmitigated Sybil routing vulnerability when claiming SHA-256 content hashes prevent DHT eclipse attacks.
*   **Reason for Rejection**: Open-data and enterprise bundles predominantly distribute over Git, HTTP/S, and S3; P2P transport belongs in standalone [`docs/rfcs/0001-p2p-sovereign-network.md`](file:///home/andreschirinos/Proyectos/open-knowledge-format/docs/rfcs/0001-p2p-sovereign-network.md) while `SPEC.md` retains only a concise profile pointer.

### Alternative 2: Embed Platform-Specific Harvesters (CKAN, Socrata, Dataverse) in Go Core

*   **Pros**: Allows `okf` CLI to harvest portal APIs without external connector scripts.
*   **Cons**: Couples the core specification and Go binary to platform-specific, versioned portal REST APIs that evolve independently of open data standards.
*   **Reason for Rejection**: Violates the core standards-only architectural boundary. Platform harvesters are isolated as external service connectors (`connectors/ckan/read_ckan.py` invoked via `okf run`).

### Alternative 3: Enforce Full OWL/RDF Triplestores for Cross-Dataset Semantic Alignment

*   **Pros**: Supports arbitrary description-logic inference and complex ontology reasoning.
*   **Cons**: Imposes high authoring complexity for human maintainers and large token overhead for AI agents compared to YAML frontmatter.
*   **Reason for Rejection**: Lightweight SKOS URI anchors (`skos`, `exact_match`, `close_match`) in YAML frontmatter provide deterministic SQL join compilation when `skos:closeMatch` is excluded from automated joins.

### Alternative 4: Ban `index.md` From Carrying Dataset Frontmatter in Single-Dataset Bundles

*   **Pros**: Enforces a single uniform rule where `index.md` never contains concept YAML keys.
*   **Cons**: Forces single-dataset repositories to create redundant wrapper markdown files alongside `index.md` and breaks existing single-dataset ODKF bundles validated by [`src/domain/validator.go`](file:///home/andreschirinos/Proyectos/open-knowledge-format/src/domain/validator.go).
*   **Reason for Rejection**: Keying Root Concept Manifest validation off explicit DCAT namespace prefixes (`type: dcat:Dataset | dcat:DataService | dcat:CatalogRecord`) in §3.1, §8, and `PE3` preserves compatibility with existing bundles while keeping shorthand `type: dataset` warned in `index.md`.

--------------------------------------------------------------------------------

## Risks & Mitigations

*   **Risk 1 — Breaking Queries That Relied on Fuzzy `skos:closeMatch` Joins**:
    *   *Impact*: Existing workflows calling `GenerateCrossDatasetJoinSQL` / `generate_cross_dataset_join_sql` on datasets linked only by `skos:closeMatch` will now receive an explicit error instead of a generated SQL string.
    *   *Mitigation*: Provide an optional `allow_close_match` boolean parameter (default `false`) for exploratory callers that explicitly opt into fuzzy joins, while keeping default automated join generation strictly exact per §14.3.1.
*   **Risk 2 — Regular Expression Backtracking (ReDoS) on Expanded Triad Grammar**:
    *   *Impact*: Expanding `triadRegex` to cover five representations plus quoted reverse-DNS segments could degrade validation latency on pathological inputs.
    *   *Mitigation*: Use Go's linear-time RE2 engine (`regexp` package) and non-overlapping character classes (`[^"]+` for quoted segments) in both Go and Python.
*   **Risk 3 — SQLite Lock Contention Under Concurrent Agent/CLI Workloads**:
    *   *Impact*: Concurrent `okf run` processes and MCP queries writing to `metadata.db` can trigger `SQLITE_BUSY` errors if connection pragmas are applied only once at schema creation.
    *   *Mitigation*: Enable persistent `PRAGMA journal_mode = WAL;` during `_init_db()` and execute per-connection `PRAGMA busy_timeout = 5000;` and `PRAGMA foreign_keys = ON;` inside `_get_connection()` on every connection instantiation.
*   **Risk 4 — v0.2 to v0.3 Bundle Version Compatibility**:
    *   *Impact*: Existing bundles declaring `okf_version: "0.2"`, `lineage.source` (singular), or `datamesh://` URIs could fail strict v0.3 validators.
    *   *Mitigation*: Accept `okf_version: "0.2"`, `lineage.source`, `datamesh://`, `.datamesh/metadata.db`, and `/.well-known/datamesh.json` as backward-compatible aliases.

## Non-Functional Requirements

*   **Security & Privacy**:
    *   Zero host script execution during `okf validate`, `okf diff`, and static `okf attest`.
    *   Mandatory URL scheme validation (`https://` / opt-in `http://`) on remote dereferencing (`OE6`) and bundle-root path containment on local paths and archive `#fragment` extractions (`OE7`).
    *   Mandatory `HAVING COUNT(*) >= min_group_size` enforcement for `zero_microdata` SQL queries (`OE3`).
*   **Backward Compatibility**:
    *   Existing bundles using `lineage.source` (singular), `datamesh://` URIs, `.datamesh/metadata.db`, and `/.well-known/datamesh.json` continue to parse as deprecated aliases without hard failures.

--------------------------------------------------------------------------------

## Detailed Implementation Plan

Work is partitioned into four sequential phases with verification gates between phases.

### Phase 1: Specification & Architecture Hardening (`open-knowledge-format`)

#### 1. [`SPEC.md`](file:///home/andreschirinos/Proyectos/open-knowledge-format/SPEC.md) (Modify)

*   **Type / Scope**: Normative Specification Document
*   **Proposed Changes**:
    *   Bump version header (line 3) and §12 (lines 835, 843, 917–928) from `0.2` to `0.3` (`okf_version: "0.3"`), clarifying `<major>.<minor>` format revisions vs. SemVer 2.0.0 dataset versions (`version: 1.0.0`).
    *   Change four outer triple-backtick fences in §10.2 (lines 655–690) and Appendix A (lines 1344–1377, 1425–1471, 1476–1513) to quadruple-backtick (` ````markdown `) fences.
    *   Fix skipped heading levels at line 19 (`## Conformance & Keyword Definitions`) and lines 1338/1379 (`## A.1 v0.1 Form`, `## A.2 v0.2 Form`).
    *   Reconcile `index.md` Root Concept Manifest rules across §3.1 (line 158), §8 (line 548), §11 (line 797), and §14.5 `PE3` (line 1074): state that `index.md` is reserved in general (`PE3-RESERVED-FILE` `WARNING` on `type: dataset | service`), except when declaring an explicit DCAT class (`type: dcat:Dataset | dcat:DataService | dcat:CatalogRecord`).
    *   Keep Part I (§11–§12) focused on Level 0 Core; move `AttestationVerdict` (`okf attest`) and `okf diff` wire contracts to Part II (`### 14.7 CLI Verification & Schema Evolution Wire Contracts`), specifying exit codes (`0` verified/no-breaking-diff, `1` refuted/major-breaking-diff, `2` runtime/parse error) and canonical SQL hashing (`NormalizeSQLCanonical`: strip SQL comments, collapse whitespace, uppercase SQL keywords, substitute `@param` / `{{ var('param') }}` bindings, SHA-256 digest).
    *   Define formal YAML schemas in §14.2 and §14.3.1:
        ```yaml
        relations:
          - rel: derived_from | supersedes | in_catalog | foreign_key | references
            target: censo_2024:ine_poblacion
        semantic_fields:
          - field_name: entidad_municipal
            concept_id: presupuesto_2025:concepts/entidad.md
            rdf_type: https://schema.org/AdministrativeArea
            unit: http://qudt.org/vocab/unit/NUM
            aggregation_behavior: sum # sum | avg | count | min | max | none
            value_labels:
              - raw_value: "GAM La Paz"
                concept_id: presupuesto_2025:concepts/gam_la_paz.md
                alt_labels: ["GAMLP", "Alcaldía de La Paz"]
        lineage:
          sources:
            - url: https://raw.source.gob.bo/censo_raw.csv
          transformations:
            - step: clean_data
              tool: scripts/clean_data.py
              hash: sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
          version: 1.0.0
          updated_at: 2026-05-01T12:00:00Z
        policy:
          profile: zero_microdata # allow_all | zero_microdata | deny
          min_group_size: 5
        policies:
          - type: odrl
            ref: https://www.w3.org/ns/odrl/2/policy.json
          - type: opa
            path: policies/privacy_guard.rego
        ```
    *   Fix §14.2 line 981 contract example (`path: https://api.gob.bo/v3/api-docs` -> `ref: https://api.gob.bo/v3/api-docs`).
    *   Update §14.5 Consumer Invariants (`OE1`–`OE7`):
        *   Update `OE3` to require `HAVING COUNT(*) >= min_group_size` (default `5`) on grouped queries under `policy: zero_microdata`.
        *   Add `OE6 (URL Scheme Guard)`: remote dereferencing MUST allow only `https://` (and opt-in `http://`), rejecting `file://`, `gopher://`, and `ftp://` schemes; add a non-normative note recommending private-IP/cloud-metadata (`169.254.169.254`) filtering for multi-tenant proxy deployments.
        *   Add `OE7 (Path Containment & Zip Slip Guard)`: local paths MUST be cleaned (`filepath.Clean`), symlink-resolved if existing on disk, and verified to remain within the bundle root; archive `#fragment` extractors MUST reject entries with `..` or leading `/`.
    *   Update §14.5 Validation Rules table (`RULE-01`–`RULE-12`):
        *   `RULE-04`: note that placing `http(s)://` in `contracts[].path` emits `PE4-REMOTE-IN-PATH` (`WARNING`).
        *   `RULE-07`: reference `§14.5`.
        *   `RULE-08`: replace colon-only regex with five-form pattern `^([a-zA-Z0-9_\-\.]+[:/][a-zA-Z0-9_\-\. ]+([:/][a-zA-Z0-9_\-\. ]+)?|(?:"[^"]+"|[a-zA-Z0-9_\-\.]+)\.[a-zA-Z0-9_\-\.]+@[a-zA-Z0-9_\-\.]+|(?:okf|odkf|datamesh)://[a-zA-Z0-9_\-\./]+|https?://[^\s#]+)(#[a-zA-Z0-9_\-\./]+)?$`.
        *   `RULE-09`: align with §7 (`<producer>/<version>`, `(agent|human|process|user|team):<id>`, `Name <email>`, `http(s)://` or `did:` URI) and uppercase `MUST`.
        *   `RULE-10`–`RULE-12`: uppercase RFC 2119 keywords (`MUST`, `SHOULD`), separate timestamps from `cadence` durations in `RULE-11` (`§5.2, §5.5`), and cite `§5.4, §14.2` in `RULE-12`.
    *   Update §7 and §14.2 signature rules: remove `verified: true` from producer frontmatter schema and specify RFC 8785 JCS canonicalization of frontmatter (excluding `signatures`) + `\n` + raw UTF-8 body bytes.
    *   Document RFC 8615 discovery at `/.well-known/okf.json` (with `/.well-known/datamesh.json` as legacy fallback) in §14.6.3.
    *   Document the `OKF_PROVENANCE_OUT` JSONL IPC event schema (`type: input | step | output | distribution | quality | metric | note | drift`, `name`, `resource`, `path`, `format`, `sha256`, `bytes`, `rows`, `cols`, `code_hash`, `severity`, `message`, `metrics`, `timestamp`) in §17.1, and replace normative `@dm.track_lineage` / `.datamesh` tokens with `@okf.track_lineage` / `.okf` (noting `datamesh` equivalents as legacy aliases).
    *   In §18.1 line 1320, replace `MUST NEVER` with RFC 2119 `MUST NOT`; in §18.2, keep abstract entity descriptions and cite [`docs/architecture/operational-state-store.md`](file:///home/andreschirinos/Proyectos/open-knowledge-format/docs/architecture/operational-state-store.md) as the normative SQLite DDL specification.

#### 2. [`docs/architecture/operational-state-store.md`](file:///home/andreschirinos/Proyectos/open-knowledge-format/docs/architecture/operational-state-store.md) (Modify)

*   **Type / Scope**: Normative Client Operational State DDL Specification
*   **Proposed Changes**:
    *   Change header line 5 from `* **Status:** Informational Architecture Specification` to `* **Status:** Normative Profile Specification (Referenced by SPEC.md §18.2)` and line 15 `MUST NEVER` to `MUST NOT`.
    *   Update line 48 comment from `/.well-known/datamesh.json` to `/.well-known/okf.json (fallback /.well-known/datamesh.json)`.
    *   Clarify in §3 (lines 130–132) that `PRAGMA foreign_keys = ON;` and `PRAGMA busy_timeout = 5000;` MUST execute on every connection instantiation, whereas `PRAGMA journal_mode = WAL;` and `PRAGMA user_version = 1;` execute during database initialization and schema migration.

*   **Phase 1 Verification Gate**: Verify zero unclosed or prematurely closed code fences in `SPEC.md` and run `go test ./...` in `src/`.

---

### Phase 2: Go Core Domain Parsers & Validation (`open-knowledge-format/src/domain`)

#### 3. [`src/domain/triad.go`](file:///home/andreschirinos/Proyectos/open-knowledge-format/src/domain/triad.go) (Modify)

*   **Type / Scope**: Go Core Domain Parser
*   **Proposed Changes**:
    *   Update `quotedAtRegex` (line 15) to `["']((?:"[^"]+"|[a-zA-Z0-9_\-\.]+)\.[a-zA-Z0-9_\-\.]+@[a-zA-Z0-9_\-\.#/]+)["']` so `ExtractTriadsFromSQL` captures quoted dataset identifiers.
    *   Update `ResourceTriad.SovereignDNS()` (lines 93–105) so that if `strings.Contains(ds, ".")`, `ds` is formatted with enclosing double quotes (`"%s"`).
    *   Update `ParseTriad` (line 113) to strip outer quotes only when the entire string starts and ends with matching quote characters (`"` or `'` or `` ` ``), preserving leading quotes on `"censo.2024".votos@bolivia`.
    *   Update reverse-DNS splitting in `ParseTriad` (lines 176–195): if `left` starts with `"`, locate the closing `"` followed by `.`, extract `left[1:closeIdx]` as dataset and `left[closeIdx+2:]` as resource, and strip any enclosing quotes in `NormalizeTriadSegment`.

#### 4. [`src/domain/validator.go`](file:///home/andreschirinos/Proyectos/open-knowledge-format/src/domain/validator.go) (Modify)

*   **Type / Scope**: Go Core Bundle & Concept Validator
*   **Proposed Changes**:
    *   Update `triadRegex` (line 61) to `^([a-zA-Z0-9_\-\.]+[:/][a-zA-Z0-9_\-\. ]+([:/][a-zA-Z0-9_\-\. ]+)?|(?:"[^"]+"|[a-zA-Z0-9_\-\.]+)\.[a-zA-Z0-9_\-\.]+@[a-zA-Z0-9_\-\.]+|(?:okf|odkf|datamesh)://[a-zA-Z0-9_\-\./]+|https?://[^\s#]+)(#[a-zA-Z0-9_\-\./]+)?$`.
    *   Add helper `ResolveBundleRoot(filePath string) string` that walks upward from `filepath.Dir(filePath)` looking for `index.md`, `okf.yaml`, or `.git` (falling back to `filepath.Dir(filePath)`).
    *   In `ValidateConceptContent` (`PE4` lines 320–333 and `RULE-06` lines 366–379):
        *   If `c.Path` (or `p.Path`) starts with `http://` or `https://`, emit `PE4-REMOTE-IN-PATH` (`SeverityWarning`, advising `ref` instead of `path`).
        *   For local paths, compute `resolved := filepath.Clean(filepath.Join(baseDir, targetRef))` (and `filepath.EvalSymlinks(resolved)` when the target exists on disk), check containment against `bundleRoot := ResolveBundleRoot(filePath)` via `filepath.Rel(bundleRoot, resolved)`, and emit `OE7-PATH-TRAVERSAL` (`SeverityError`) if `strings.HasPrefix(rel, "..") || filepath.IsAbs(rel)`.

#### 5. [`src/domain/validator_test.go`](file:///home/andreschirinos/Proyectos/open-knowledge-format/src/domain/validator_test.go) (Modify)

*   **Type / Scope**: Go Domain Unit Tests
*   **Proposed Changes**:
    *   Add unit tests for `ParseTriad`, `SovereignDNS()`, `ExtractTriadsFromSQL`, and `ValidateTriad` with `"censo.2024".votos@bolivia#seccion.csv`, `odkf://`, and `https://` URLs.
    *   Add unit tests for `OE7-PATH-TRAVERSAL` (`SeverityError` on `../../../etc/passwd`) and `PE4-REMOTE-IN-PATH` (`SeverityWarning` on `path: https://...`).

*   **Phase 2 Verification Gate**: Run `go test ./domain/...` inside `/home/andreschirinos/Proyectos/open-knowledge-format/src`.

---

### Phase 3: Go Core Use Cases, Manifests & Attestation (`open-knowledge-format/src`)

#### 6. [`src/usecases/publish_and_semantic_usecase.go`](file:///home/andreschirinos/Proyectos/open-knowledge-format/src/usecases/publish_and_semantic_usecase.go) (Modify)

*   **Type / Scope**: Go Core Semantic Alignment Use Case
*   **Proposed Changes**:
    *   In `GenerateCrossDatasetJoinSQL` (lines 398–415), filter `correspondences` to exclude entries where `corr.MatchPredicate == "skos:closeMatch"`.
    *   If the filtered slice is empty, return an explicit error (`fmt.Errorf("no exact or verified semantic correspondences found between %s and %s (skos:closeMatch excluded per SPEC §14.3.1)", sourceMap.FieldName, targetMap.FieldName)`) rather than falling back to raw column equality.

#### 7. [`src/adapters/outbound/manifests/llmstxt_parser.go`](file:///home/andreschirinos/Proyectos/open-knowledge-format/src/adapters/outbound/manifests/llmstxt_parser.go) (Modify)

*   **Type / Scope**: Go Core `llms.txt` Manifest Parser
*   **Proposed Changes**:
    *   Update `domainRegex` (line 17) to `regexp.MustCompile(` \((?:Dominio|Domain):\s*([^)]*?)(?:\.|\)|(?:Recursos|Resources):)`)` and `recRegex` (line 18) to `regexp.MustCompile(`(?:Recursos|Resources):\s*([^)]+)\)` )`.
    *   Update description truncation (line 197) to strip both `(Dominio:` and `(Domain:` suffixes from `cleanDesc`.

#### 8. [`src/usecases/diff_and_attest.go`](file:///home/andreschirinos/Proyectos/open-knowledge-format/src/usecases/diff_and_attest.go) (Modify)

*   **Type / Scope**: Go Core `okf diff` & `okf attest` Use Case
*   **Proposed Changes**:
    *   In `extractSanctionedComputation` (lines 404–418), replace the raw `strings.Contains(compPath, "..")` rejection with bundle-root containment using `domain.ResolveBundleRoot(baseDir)`: compute `fullPath := filepath.Clean(filepath.Join(baseDir, filepath.FromSlash(compPath)))`, verify `rel, err := filepath.Rel(bundleRoot, fullPath)` does not start with `".."` and is not absolute, and reject only when the path escapes `bundleRoot`.

#### 9. [`src/usecases/publish_and_semantic_usecase_test.go`](file:///home/andreschirinos/Proyectos/open-knowledge-format/src/usecases/publish_and_semantic_usecase_test.go) & [`src/adapters/outbound/manifests/manifests_test.go`](file:///home/andreschirinos/Proyectos/open-knowledge-format/src/adapters/outbound/manifests/manifests_test.go) (Modify)

*   **Type / Scope**: Go Use Case & Manifest Unit Tests
*   **Proposed Changes**:
    *   In `publish_and_semantic_usecase_test.go`: add tests verifying `skos:closeMatch` exclusion and fail-fast error on empty exact matches in `GenerateCrossDatasetJoinSQL`, and `../computations/revenue.sql` relative path resolution in `AttestComputation`.
    *   In `manifests_test.go`: add tests verifying English `(Domain: Finance. Resources: res1, res2)` extraction and description truncation.

*   **Phase 3 Verification Gate**: Run `go test ./...` inside `/home/andreschirinos/Proyectos/open-knowledge-format/src`.

---

### Phase 4: Python SDK Alignment (`../datamesh-sdk/bindings/python`)

*Note: Requires sibling repository `/home/andreschirinos/Proyectos/datamesh-sdk` with `.venv/bin/python`.*

#### 10. [`../datamesh-sdk/bindings/python/datamesh/usecases/semantic_alignment.py`](file:///home/andreschirinos/Proyectos/datamesh-sdk/bindings/python/datamesh/usecases/semantic_alignment.py) (Modify)

*   **Type / Scope**: Python SDK Semantic Alignment Use Case
*   **Proposed Changes**:
    *   Add optional `allow_close_match: bool = False` parameter to `generate_cross_dataset_join_sql` (lines 222–242); filter out `c.match_predicate == "skos:closeMatch"` when `allow_close_match` is `False`, and raise `ValueError` if zero correspondences remain.

#### 11. [`../datamesh-sdk/bindings/python/datamesh/domain/models.py`](file:///home/andreschirinos/Proyectos/datamesh-sdk/bindings/python/datamesh/domain/models.py) & [`../datamesh-sdk/bindings/python/datamesh/adapters/catalog/llmstxt_catalog.py`](file:///home/andreschirinos/Proyectos/datamesh-sdk/bindings/python/datamesh/adapters/catalog/llmstxt_catalog.py) (Modify)

*   **Type / Scope**: Python SDK Domain & Catalog Parser
*   **Proposed Changes**:
    *   In `CanonicalURI.parse` (`models.py` lines 76, 130–136): unwrap outer quotes only when paired (`(s[0] == s[-1]) and s[0] in "'\"`"`), and in reverse-DNS `<dataset>.<resource>@<catalog>`, if `left` starts with `"`, split on the `.` immediately following the closing `"` and unquote `dataset`.
    *   In `llmstxt_catalog.py` (lines 18–19, 109–115): update `domain_pattern` to `r"\((?:Dominio|Domain):\s*([^)]*?)(?:\.|\)|(?:Recursos|Resources):)"`, `rec_pattern` to `r"(?:Recursos|Resources):\s*([^)]+)\)"`, and truncate `clean_desc` on both `(Dominio:` and `(Domain:`.

#### 12. [`../datamesh-sdk/bindings/python/datamesh/constants.py`](file:///home/andreschirinos/Proyectos/datamesh-sdk/bindings/python/datamesh/constants.py) & [`../datamesh-sdk/bindings/python/datamesh/adapters/storage/sqlite_lineage.py`](file:///home/andreschirinos/Proyectos/datamesh-sdk/bindings/python/datamesh/adapters/storage/sqlite_lineage.py) (Modify)

*   **Type / Scope**: Python SDK Operational State Store
*   **Proposed Changes**:
    *   In `constants.py` and `sqlite_lineage.py`, resolve metadata DB path in strict precedence order: (1) `OKF_METADATA_DB`, (2) `DATAMESH_METADATA_DB`, (3) `$OKF_HOME/metadata.db`, (4) `$DATAMESH_HOME/metadata.db`, (5) `./.okf/metadata.db` (if `.okf` exists), (6) `./.datamesh/metadata.db` (if `.datamesh` exists), (7) `~/.okf/metadata.db` (if `~/.okf` exists), (8) `~/datamesh/metadata.db`.
    *   In `sqlite_lineage.py`: execute `PRAGMA busy_timeout = 5000;` alongside existing `PRAGMA foreign_keys = ON;` in `_get_connection()`, and execute `PRAGMA journal_mode = WAL;` and conditional `PRAGMA user_version = 1;` migration in `_init_db()`.

#### 13. [`../datamesh-sdk/bindings/python/tests/test_publishing_and_semantics.py`](file:///home/andreschirinos/Proyectos/datamesh-sdk/bindings/python/tests/test_publishing_and_semantics.py) (Modify)

*   **Type / Scope**: Python SDK Unit Tests
*   **Proposed Changes**:
    *   Add tests verifying `generate_cross_dataset_join_sql` excludes `skos:closeMatch` and raises `ValueError` when only `closeMatch` links exist, `CanonicalURI.parse('"censo.2024".votos@bolivia#divisas.csv')` resolves `dataset == "censo.2024"`, and English `llms.txt` entries extract `domain`, `resources`, and clean descriptions.

### Verification Plan

*   **Go Compilation & Unit Tests**: Run `go test ./...` inside `/home/andreschirinos/Proyectos/open-knowledge-format/src` and verify 100% pass rate.
*   **Python Unit Tests**: Run `/home/andreschirinos/Proyectos/datamesh-sdk/.venv/bin/python -m pytest /home/andreschirinos/Proyectos/datamesh-sdk/bindings/python/tests -q` and verify 100% pass rate.
*   **Specification Structural Audit**: Verify zero unclosed or prematurely closed code fences in `SPEC.md`.
