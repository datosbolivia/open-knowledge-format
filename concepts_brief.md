# Comprehensive Architecture & Specification Review: Open Knowledge Format (SPEC.md v0.2)

* **Author:** andreschirinos
* **Date:** 2026-11-11

## Overview

This brief synthesizes an exhaustive multi-lens audit of the Open Knowledge Format specification ([`SPEC.md`](file:///home/andreschirinos/Proyectos/open-knowledge-format/SPEC.md)) across seven independent specialized review lenses: System Design & Attack Surfaces (`egm-design`), Codebase Drift vs. Go Core (`egm-codebase-details`), Skeptical Minimalism & Complexity Budget (`egm-zoom-out`), Naive Reader Comprehension (`egm-comprehension`), Implementation Readiness & Determinism (`egm-implementation-readiness`), Structural Layout & Formatting (`egm-structural`), and Inventive Contradiction Resolution (`egm-inventive-catalyst`).

The core value proposition of OKF is a human- and agent-readable meta-standard that combines markdown documentation, frictionless schema referencing, epistemic trust metadata, parameter-constrained attested computations, and canonical triad addressing without requiring heavy catalog infrastructure. However, the current specification suffers from severe architectural divergence: Part I presents a lightweight, zero-tooling document format, while Part II attempts to specify a decentralized P2P operating system (Kademlia DHT, libp2p, custom DNS records) and leaks downstream SDK implementation details (SQLite schemas, WAL PRAGMAs, Hexagonal OOP ports). Furthermore, critical security vulnerabilities (unsandboxed remote code execution in `okf attest`), cryptographic category errors (claiming SHA-256 prevents Sybil attacks in DHT discovery), and drift between `SPEC.md` and the Go reference engine require immediate rectification.

## Analysis

### 1. The Identity Split: Core Document Standard vs. Speculative P2P Operating System
The specification oscillates between two incompatible scopes:
- **Part I (§1–§13):** A clean, static, tool-agnostic metadata convention ("if you can `cat` a file, you can read OKF").
- **Part II (§14–§18):** An aspirational distributed networking runtime specifying Kademlia DHT 256-bit lookups, libp2p GossipSub topics, ICE/STUN/AutoNAT hole punching, Circuit Relay v2, custom sovereign resource records (`SOA`, `CAT`, `RES`, `SCH`, `LIN`, `SRV`), and downstream database tables.

**Findings & Risks:**
- **Zero Real-World Adoption for DHT (§15):** 99.9% of enterprise, academic, and open-data publishers use HTTP/S, Git, and S3/MinIO. Git already provides cryptographic, decentralized, content-addressed distribution. Specifying an unproven P2P network stack inside a document standard raises implementation friction to near-impossible heights.
- **Leaking SDK Internals (§16, §17.2, §18):** Section 16 prescribes internal Go/Python Hexagonal Architecture ports (`StoragePort`, `QueryEnginePort`), and §18 dictates private SQLite connection PRAGMAs (`WAL`, `busy_timeout = 5000`) and table DDL (`cache_entries`, `executions`). Format specifications define syntax and wire invariants, not the private caching layers of specific client runtimes (`datamesh-sdk`).
- **Preamble Inconsistency:** Section 1 promises "no required query engine or storage infrastructure", directly contradicted by §16.1 and §18.2.

### 2. Critical Security Vulnerabilities & Cryptographic Flaws
- **Unsandboxed Consumer Remote Code Execution (SEC-01 / Blocker):** Section 10.2 specifies that attested computations execute arbitrary producer-supplied scripts on the consumer machine (`attester: { resource: references/attesters/revenue.py }`, line 624: *"It is meant to run consumer-side"*). Section 12.3 admits that attester sandboxing ABI was deferred. Consuming an untrusted OKF bundle and running `okf attest` executes arbitrary Python code directly on the host system without process isolation.
- **Sybil / Eclipse Vulnerability in Kademlia DHT (NET-01 / Blocker):** Section 15.5.1 claims that Sybil attacks cannot occur because content is hashed with SHA-256. This is a severe cryptographic category error. Content hashing (CIDs) verifies payload integrity *after* retrieval, but does not protect DHT *routing and discovery*. In an open Kademlia DHT lacking Proof-of-Work, staking, or PKI, an attacker generates thousands of node IDs near `SHA-256("okf:catalog:" + catalog)`, eclipses target routing tables, and censors or hijacks provider records at will.
- **Unauthenticated String Trust Tiers (SEC-02 / Major):** Sections 5.3 and 7 determine "human-reviewed" trust status solely by checking whether the unauthenticated string `verified[].by` starts with `human:`. Any agent or adversary can inject `verified: { by: human:auditor }`. Trust tier elevation must be bound to cryptographic signatures (`did:github`, Ed25519).

### 3. State & Lineage Contradictions
- **Episodic State Contradiction (DAT-01 / Major):** Section 18.1 explicitly defines `metadata.db` as a purgeable, disposable operational cache ("if metadata.db is purged or lost, the runtime self-heals by re-probing endpoints"). Yet §17.2 and §18.2 define `executions`, `lineage_sources`, `lineage_transformations`, and `artifacts` exclusively inside `metadata.db`. If the cache is purged, all un-exported operational execution lineage is irreversibly destroyed.
- **SQLite Lock Contention (CONC-01 / Minor):** Multi-agent workloads executing queries, running scripts via `okf run`, and accessing MCP tools concurrently will encounter unhandled `SQLITE_BUSY` crashes due to the single-file database lock and 5000ms timeout.
- **Missing Foreign Keys & Migrations:** SQLite foreign keys are disabled by default. Section 18.3 prescribes WAL mode but omits `PRAGMA foreign_keys = ON;`, rendering cascading deletes inert. No schema migration strategy (`PRAGMA user_version`) is specified.

### 4. Implementation Drift: SPEC.md vs. Reference Go Core
The reference engine (`src/`) has evolved ahead of `SPEC.md`:
- **Missing Validation Rules (RULE-09 to RULE-12):** `SPEC.md` §14.5 ends at `RULE-08`. The Go validator actively enforces `RULE-09-ACTOR-FORMAT`, `RULE-10-STATUS-UNKNOWN`, `RULE-10-SUPERSEDED-TARGET`, `RULE-11-EXPIRES-SYNTAX`, `RULE-11-CADENCE-SYNTAX`, `RULE-11-GENERATED-AT-SYNTAX`, and `RULE-12-DAG-CYCLE` / `RULE-12-CIRCULAR-REFERENCE`.
- **Severity Inversions:** `SPEC.md` states missing local contracts in RULE-04 are `ERROR`; code emits `WARNING` (`PE4-CONTRACT-LOCAL-NOT-FOUND`). `SPEC.md` states empty bundles in RULE-07 are `WARNING`; code emits `ERROR` (`RULE-07-EMPTY-BUNDLE`).
- **Addressing Forms (3 vs. 5) & Fragment Parsing:** `SPEC.md` §14.4 lists only 3 addressing forms and its RULE-08 regex rejects `#fragment`. The Go engine (`triad.go`) implements 5 forms (colon, slash, reverse DNS, sovereign URI, direct HTTP URL) and parses `#fragment` to extract subfiles from ZIP archives in memory.
- **Access Policy Gaps:** `policy.go` requires `HAVING COUNT(*) >= min_group_size` for grouped SQL queries and permits `admin`/`analyst`/`official` bypass of `zero_microdata`, neither of which is documented in `SPEC.md`.
- **Decoupled CKAN Architecture:** Section 16.1 still depicts `CKAN API` as a core catalog adapter. In reality, CKAN was completely removed from the Go core and relocated to the decoupled external service connector [`connectors/ckan/`](file:///home/andreschirinos/Proyectos/open-knowledge-format/connectors/ckan/).
- **Undocumented Cryptographic Entities:** `src/domain/crypto.go` and `dataproduct.go` implement `did:github:<username>` resolution against OpenSSH keys and frontmatter `signatures[]`, both missing from `SPEC.md`.

### 5. Multi-Standard Bloat & Theoretical Translation Guarantees
- **Excessive Standards Coverage:** The specification claims normative support for Frictionless, DCAT v3, OpenAPI 3.1, Arrow Flight, CWL, PROV-O, ODRL 2.2, OPA/Rego, DPV, XACML, SDMX, DDI, STAC, and ISO 19115.
- **The "Near-Lossless Translation" Fallacy:** Section 16.4 claims near-lossless roundtrip translation between DCAT-AP (RDF open-world graph) and Frictionless/Arrow (closed-world physical tabular schemas). RDF blank nodes, reification, and arbitrary multi-inheritance cannot roundtrip into tabular JSON without losing semantics or dumping into untyped extensions.
- **SKOS Join Explosion:** Section 14.3.1 proposes automated cross-dataset SQL `JOIN` generation by traversing `skos:closeMatch` to Wikidata URIs. In practice, `closeMatch` is non-transitive and lacks cardinality constraints, leading to Cartesian explosions.

### 6. TRIZ Inventive Contradictions & Resolutions
1. **Human-Readability vs. Machine-Parsability:**
   - *False Compromise:* Prose markdown headers double as operational contracts, forcing parsers to use heuristic regex or LLM extraction.
   - *Inventive Resolution (Separation in Space / Preliminary Action):* Enforce a dual-faceted envelope. The frontmatter carries the strict, typed AST or contract hash; pre-commit/compile-on-save generators derive human-readable markdown summary tables deterministically.
2. **Static Simplicity vs. Analytical Performance:**
   - *False Compromise:* Either passive static files (no compute) or heavy server infrastructure (DuckDB Flight SQL daemons).
   - *Inventive Resolution (Segmentation & Dynamics):* Static Parquet/DuckDB assets publish lightweight index footers (min/max, Bloom filters, byte offsets) alongside static files. Clients execute analytical SQL using zero-copy HTTP Range requests (RFC 7233) via in-process engines (DuckDB-WASM / embedded DuckDB) without server daemons.
3. **Sovereign Federation vs. Global Coherence:**
   - *False Compromise:* Tolerate broken links and manual SQL case-expressions across colliding namespaces.
   - *Inventive Resolution (Intermediary & Self-Service):* Decouple human sovereign aliases (`catalog:dataset:resource`) from data schemas via cryptographic Content Identifiers (CIDs). Identical schemas across independent publishers collapse to the same Merkle node.
4. **Meta-Standard Flexibility vs. Deterministic Validation:**
   - *False Compromise:* Tolerate arbitrary keys (degrading validation to YAML syntax checks) or bloat core with monolithic validators for dozens of standards.
   - *Inventive Resolution (Inversion of Control):* External contracts reference content-addressed micro-validators (WASM binaries or JSON Schemas). The core engine validates arbitrary standards via the uniform attestation wire format without native grammar dependencies.

## Goals vs. Non-Goals

### In-Scope Goals
* **P2P Decoupling:** Quarantine §15 (Kademlia DHT, custom sovereign DNS records, GossipSub NAT traversal) out of `SPEC.md` into an experimental standalone RFC (`docs/rfcs/0001-p2p-transport.md`).
* **SDK Leakage Removal:** Move §16 (Hexagonal OOP ports), §17.2, and §18 (SQLite schema DDL and connection PRAGMAs) into `datamesh-sdk` internal architecture documentation.
* **Codebase Alignment:** Synchronize `SPEC.md` with reference Go implementation: document RULE-09 through RULE-12, reconcile RULE-04 and RULE-07 severities, document 5-form triad addressing and `#fragment` support, document `did:github` and frontmatter `signatures[]`, and update §16.1 to reflect CKAN as an external service connector.
* **Security & Attestation Hardening:** Replace unsandboxed host script execution in `okf attest` with sandboxed execution (WASM or Docker container boundary) or declarative query receipts. Formally mandate cryptographic signature verification for elevated trust tiers.
* **Standards Pruning:** Focus normative contracts on three core pillars: Frictionless Table Schema (tabular), OpenAPI 3.1 (services), and SQL/dbt (transformations). Downgrade CWL, ODRL, DPV, XACML, SDMX, and DDI to non-normative community extension profiles.
* **Structural Harmonization:** Introduce explicit `# Part I — Core OKF` container, move Appendix A to the end of the document, fix H3/H4 nesting, translate Spanish table terms to English, add explicit language tags to code fences, and include the BCP 14 / RFC 2119 clause in the preamble.

### Out-of-Scope Non-Goals
* Re-implementing a full RDF/OWL ontology engine inside OKF.
* Guaranteeing lossless bidirectional translation between arbitrary external standards.
* Supporting centralized distributed database clustering inside `open-knowledge-format`.

## Alternatives Weighed

* **Alternative 1: Retain P2P Kademlia DHT in Core Standard (§15).**
  - *Rejected:* Adds massive speculative complexity (~400 lines) with zero production adoption. Git, S3, and static HTTP web servers satisfy 99.9% of real-world decentralized open data distribution requirements without NAT traversal or Sybil vulnerabilities.
* **Alternative 2: Embed Harvesters for Every Platform (CKAN, Socrata, Dataverse) in Go Core.**
  - *Rejected:* Violates the core standards-only architectural principle. Software platforms evolve proprietary APIs. Handling them as decoupled service connectors (`connectors/ckan/`) with OpenAPI and CLI contracts keeps the core engine minimal and pure.
* **Alternative 3: Enforce Native OWL/RDF Triplestores for Semantic Alignment.**
  - *Rejected:* Heavy semantic triplestores impose unacceptable cognitive overhead on humans and token costs on LLM agents. Lightweight SKOS URI anchors in frontmatter achieve 90% of cross-catalog disambiguation at a fraction of the complexity.

## Key Decisions & Staging for Next Steps

1. **Stage 1: Structural & Normative Cleanup (Immediate):**
   - Add RFC 2119 preamble and `# Part I — Core OKF` heading.
   - Relocate Appendix A after Part II.
   - Harmonize heading levels, fix language consistency in Table 7, tag code fences.
   - Add RULE-09 to RULE-12 to the normative validation table (§14.5) and fix RULE-04/RULE-07 severity inconsistencies.
   - Document the 5 canonical triad addressing forms and `#fragment` syntax.
2. **Stage 2: Scope Decoupling & Modularization:**
   - Move §15 (P2P/DHT) to `docs/rfcs/p2p-sovereign-network.md`.
   - Move §16 (Hexagonal ports) and §18 (SQLite DDL) to `datamesh-sdk` documentation.
   - Update §16.1 architecture diagram to show external service connectors rather than core CKAN adapters.
3. **Stage 3: Security & Trust Boundary Hardening:**
   - Specify sandboxed execution requirements for `okf attest`.
   - Document `did:github` OpenSSH public key verification and frontmatter `signatures[]`.
   - Clarify that string prefixes (`human:`) provide provenance metadata, not cryptographic trust verification.
