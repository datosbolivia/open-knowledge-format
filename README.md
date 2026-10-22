# Open Knowledge Format (OKF)

### 📖 [Read the Open Knowledge Format v0.2 specification → SPEC.md](SPEC.md)

> **This repository is primarily about the [Open Knowledge Format
> (OKF)](SPEC.md).**
>
> OKF is a **universal, vendor-neutral format** for representing knowledge
> as plain markdown files with YAML frontmatter. It is **not tied to any
> particular agent, framework, model provider, or serving system**. The
> goal is simple:
>
> - **Anyone can produce** OKF — humans authoring by hand, agents built on
>   any framework (Google ADK, LangChain, custom), export pipelines from
>   existing catalogs (Dataplex, Unity Catalog, Collibra, …), or scripts
>   walking a database.
> - **Anyone can serve and consume** OKF — a static file server, a
>   knowledge-management UI (Obsidian, Notion, MkDocs), an LLM loading
>   files into context, a search index, or a graph viewer like the one
>   bundled in this repo.
>
> The specification is complemented by a **native Domain-Driven Design (DDD) Go core**
> in [`src/`](src/), providing high-performance validation, bundle inspection,
> and self-contained interactive Cytoscape.js visualization without external cloud or framework dependencies.
>
> **See OKF in practice** — three ready-to-browse bundles produced by this
> agent, checked into [`bundles/`](bundles/):
>
> - [`bundles/ga4/`](bundles/ga4/) — GA4 e-commerce dataset
>   ([viz.html](bundles/ga4/viz.html))
> - [`bundles/stackoverflow/`](bundles/stackoverflow/) — Stack Overflow
>   public dataset ([viz.html](bundles/stackoverflow/viz.html))
> - [`bundles/crypto_bitcoin/`](bundles/crypto_bitcoin/) — Bitcoin
>   blocks/transactions ([viz.html](bundles/crypto_bitcoin/viz.html))
> - [`bundles/acme_retail/`](bundles/acme_retail/) — Acme Retail
>   ([viz.html](bundles/acme_retail/viz.html))
> - [`bundles/air_quality_bolivia/`](bundles/air_quality_bolivia/) — Red MoniCA Calidad del Aire Bolivia (Parquet + CSV + SKOS)
> - [`bundles/cndc_despacho_electrico/`](bundles/cndc_despacho_electrico/) — Despacho Eléctrico Dinámico SIN CNDC (DataService + PROV-O)
> - [`bundles/censo_2024_cartografia/`](bundles/censo_2024_cartografia/) — Cartografía Censal INE 2024 (Recurso Compuesto ZIP + Fragment Identifiers)
> - [`bundles/udabol_salud_syndicated/`](bundles/udabol_salud_syndicated/) — Vigilancia Epidemiológica Dengue UDABOL (Sindicación Multicatálogo + DNS Soberano)

## Why OKF?

OKF represents catalog knowledge as plain markdown files with YAML
frontmatter, organized in a directory hierarchy. That choice unlocks a few
properties that are hard to get from a service-owned metadata store:

- **Human- and agent-readable.** No SDK or query language stands between a
  reader and the content. An engineer can `cat` a concept; an LLM can ingest
  it verbatim into context.
- **Version-controllable out of the box.** Bundles live in git. Pull
  requests, line-by-line diffs, blame, and review workflows just work —
  knowledge curation becomes a normal software-engineering activity.
- **Portable and lock-in free.** A bundle is a directory. Ship it as a
  tarball, host it in any repo, mount it from any filesystem, or sync it to
  any system that speaks files. No proprietary API stands between you and
  your metadata.
- **Mixes structured and unstructured data deliberately.** Use frontmatter
  for the few fields you want to query, filter, or index on (`type`,
  `resource`, `tags`, `generated`, `status`); use the markdown body for the
  prose, schemas, and example queries that LLMs and humans actually read.
- **Trust, provenance, and freshness are first-class.** v0.2 puts queryable
  signals in frontmatter — where a concept came from (`sources` with per-source
  credibility signals), who produced and confirmed it (`generated`, `verified`,
  from which consumers derive a trust tier), and whether it is still current
  (`status`, `stale_after`) — so an agent-maintained corpus stays trustable
  without any bespoke runtime.
- **Minimally opinionated, freely extensible.** A small set of required
  keys ensures interoperability, but bundles can carry arbitrary extra
  frontmatter keys and arbitrary body sections without breaking
  consumers.
- **Composes with existing tooling.** Many knowledge tools — Notion,
  Obsidian, MkDocs, Hugo, Jekyll — already speak markdown plus YAML
  frontmatter, so bundles can be browsed, edited, or rendered without
  custom UI.
- **Progressive disclosure built in.** Auto-generated `index.md` files
  let an agent or human navigate the hierarchy one level at a time
  instead of loading the entire bundle into context.
- **Graph-shaped, not just tree-shaped.** Concepts link to each other via
  normal markdown links, expressing relationships richer than the
  parent/child implied by the directory layout.

The net effect is that reference agents, consumption agents, and humans
collaborate on the same artifacts in the same way they already collaborate
on source code.

## Install

```
python3.13 -m venv .venv
.venv/bin/pip install --index-url https://pypi.org/simple/ -e .[dev]
```

## Credentials

- BigQuery: `gcloud auth application-default login` plus a project for billing
  (`gcloud config set project <id>`). Public datasets are readable, but the
  caller's project is billed for query bytes.
- Gemini: set `GEMINI_API_KEY` (AI Studio) **or** use Vertex AI by setting
  `GOOGLE_GENAI_USE_VERTEXAI=true`, `GOOGLE_CLOUD_PROJECT=<id>`, and
  `GOOGLE_CLOUD_LOCATION=<region>`.

## How the reference agent works

The reference agent runs in two passes. The **BQ pass** writes one OKF
doc per concept the source advertises, using BigQuery metadata alone.
The **web pass** then runs the LLM as its own crawler: it receives a
list of seed URLs (provided via `--web-seed` or `--web-seed-file`),
fetches the seeds via the `fetch_url` tool, and decides which outbound
links are worth following based on whether they look like authoritative
documentation for the existing concepts. For each page it fetches, the
agent chooses to (a) enrich one or more existing concept docs, (b) mint
a standalone `references/<slug>` doc, or (c) skip. A hard
`--web-max-pages` cap and a same-domain allowed-hosts filter
(configurable via `--web-allowed-host`) are enforced inside the tool,
so the agent cannot overrun. Use `--no-web` to skip the web pass.

## Run

Minimum invocation — point at a BigQuery dataset and a bundle output
directory. Seeds for the web pass are explicit; omit them (or pass
`--no-web`) to run BQ-only:

```
.venv/bin/python -m reference_agent enrich \
    --source bq \
    --dataset <project>.<dataset> \
    --web-seed-file <path/to/seeds.txt> \
    --out ./bundles/<name>
```

Iterate on a single concept by adding `--concept <type>/<name>` (e.g.
`--concept tables/events_`); repeatable.

## Samples

Each sample pairs a **recipe** (`samples/<name>/`, with the seed URLs and
exact `enrich` command) with the **produced bundle** (`bundles/<name>/`)
that the recipe generated. Open the recipe to reproduce; open the bundle
to browse the result directly.

- **GA4 Google Merchandise Store** — public e-commerce dataset, seeded
  with canonical GA4 BigQuery Export documentation URLs.
  · [recipe](samples/ga4_merch_store/README.md)
  · [bundle](bundles/ga4/)
  · [viz.html](bundles/ga4/viz.html)
- **Stack Overflow** — public dataset (mirror of the Stack Exchange Data
  Dump), seeded with the community's canonical schema references.
  Exercises multi-concept enrichment from cross-cutting docs pages.
  · [recipe](samples/stackoverflow/README.md)
  · [bundle](bundles/stackoverflow/)
  · [viz.html](bundles/stackoverflow/viz.html)
- **Bitcoin (crypto)** — public dataset (blocks, transactions, inputs,
  outputs) from the `bitcoin-etl` pipeline. Exercises cross-table
  foreign-key relationships in prose.
  · [recipe](samples/crypto_bitcoin/README.md)
  · [bundle](bundles/crypto_bitcoin/)
  · [viz.html](bundles/crypto_bitcoin/viz.html)

## Visualize

The `visualize` subcommand renders any OKF bundle as a **self-contained
interactive HTML file** — one file, no backend, no install on the
viewing side. Open it in any modern browser, share it as an artifact,
host it on a static file server, or commit it next to the bundle (as
this repo does).

The viewer is itself a proof-of-concept *consumer* of OKF, mirroring
the way the reference agent is a proof-of-concept *producer*. OKF
bundles can be consumed by anything that reads markdown; this is just
one shape.

### What it shows

- A **force-directed graph** of every concept in the bundle, with
  colored nodes by type (datasets, tables, references, …) and directed
  edges drawn from each cross-link in the markdown bodies.
- A **detail panel** for the selected concept showing its frontmatter
  (description, resource link, tags) and its rendered markdown body —
  with internal `[…](/path/to/concept.md)` links rewired to navigate
  within the viewer instead of following the path.
- A **"Cited by" backlinks** list under each concept (computed from the
  reverse of the link graph).
- A **search box** (matches title, concept id, and tags), a **type
  filter**, and switchable graph layouts (cose / concentric /
  breadth-first / circle / grid).

### Generate

```
./src/bin/okf visualize --bundle ./bundles/<name>
```

That writes `bundles/<name>/viz.html`. Flags:

| Flag           | Default                | Description                                 |
|----------------|------------------------|---------------------------------------------|
| `--bundle`     | *(required)*           | Bundle root directory.                      |
| `--out`        | `<bundle>/viz.html`    | Output HTML path.                           |
| `--name`       | bundle directory name  | Display name shown in the viewer header.    |

Example, writing the output somewhere else and overriding the header:

```
./src/bin/okf visualize \
    --bundle ./bundles/crypto_bitcoin \
    --out /tmp/btc.html \
    --name "Bitcoin OKF"
```

### How it's built

The HTML embeds the bundle as a JSON blob and uses
[Cytoscape.js](https://js.cytoscape.org/) for the graph and
[marked](https://marked.js.org/) for in-browser markdown rendering,
both loaded from a CDN. No data leaves the page; the bundle is parsed
once at generation time and serialized into the file.

## Specification-Driven Design (SDD) & Formal Architecture

The Open Knowledge Format follows a strict **Specification-Driven Design (SDD)** methodology:
* **Normative SDD Document:** [`docs/SPECIFICATION_DRIVEN_DESIGN.md`](docs/SPECIFICATION_DRIVEN_DESIGN.md)
* **Formal Grammars (EBNF):** Reverse Sovereign DNS (`<dataset>.<resource>@<catalog>`), Canonical URIs (`okf://`), Triads, and Archive Fragment Identifiers (`#subpath`).
* **Abstract Hexagonal Ports:** `StoragePort` (RFC 7233 Range-reads), `CatalogResolverPort`, `DataProductResolverPort`, `QueryEnginePort`, `TrustVerifierPort`.
* **State Machines (FSM):** Formal state transitions for recursive catalog crawling ($MAX\_DEPTH=8$), HTTP Range zip extraction, and dynamic DataService failover.
* **Executable Conformance Test Suite (ECTS):** Golden test vectors to certify downstream client runtimes (e.g. [`datamesh-sdk`](file:///home/andreschirinos/Proyectos/datamesh-sdk)).
* **Problem Specifications:** Detailed RFC-style technical specs in [`docs/problemas/`](docs/problemas/README.md).

## Core Go CLI & Testing

The format engine and CLI are built with Go 1.22+:

```bash
cd src
go build -o bin/okf ./cmd/okf

# Run all domain and use case tests
go test -v ./...

# Validate an OKF concept or bundle
./bin/okf validate ../bundles/air_quality_bolivia/index.md

# Generate interactive visualization
./bin/okf visualize --bundle ../bundles/air_quality_bolivia --out /tmp/air_viz.html
```
