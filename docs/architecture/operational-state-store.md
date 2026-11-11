# OKF Client Operational State Architecture (`metadata.db`)

* **Target Implementation:** Downstream Client Engines & SDKs (`datamesh-sdk`)
* **Reference Implementation:** Embedded SQLite State Store
* **Status:** Informational Architecture Specification

---

## 1. Storage Partitioning Principle (Human Intent vs. Operational State)

An OKF client runtime MUST maintain a strict boundary between declarative human intent and machine-generated operational state:

1. **Declarative User Configuration (`okf.yaml` / `trust.yaml`):**
   - Explicit human intents, including catalog seed URLs, network timeouts, local workspace roots, query engine preferences (`duckdb`, `inmem`), and *Root Trust Anchors* (cryptographic author allowlists, OpenSSH public key fingerprints, verification policies `strict` vs `permissive`), MUST reside in human-readable, version-controlled YAML/JSON text files.
   - Deleting or rebuilding operational state MUST NEVER alter human configuration files.
2. **Operational State & Persistent Lineage:**
   - Client runtimes manage local caching and execution lineage via an embedded SQLite database (`$OKF_HOME/metadata.db` or `.okf/metadata.db`).
   - **Architectural Boundary:** Runtimes MUST distinguish between:
     - **Disposable Caches (`cache_entries`, `dns_cache`, `keys_cache`):** Ephemeral data that can be safely purged. If deleted, the runtime self-heals by re-probing endpoints, re-fetching keys, and re-downloading resources.
     - **Persistent Lineage Ledger (`executions`, `lineage_sources`, `lineage_transformations`, `artifacts`):** The local cryptographic audit trail of computations and transformations. Runtimes SHOULD export this lineage to `log.md` or PROV-O manifests before purging local databases.

---

## 2. Relational Schema of `metadata.db`

```sql
-- 1. Downloaded Resource Cache
CREATE TABLE IF NOT EXISTS cache_entries (
    key TEXT PRIMARY KEY,
    original_uri TEXT NOT NULL,
    local_path TEXT NOT NULL,
    format TEXT NOT NULL,
    size_bytes INTEGER NOT NULL,
    cached_at TEXT NOT NULL,
    etag TEXT,
    sha256 TEXT,
    source_service TEXT,
    update_frequency TEXT,
    last_accessed TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_cache_uri ON cache_entries(original_uri);
CREATE INDEX IF NOT EXISTS idx_cache_path ON cache_entries(local_path);

-- 2. Sovereign Endpoint & DNS Probing Cache
CREATE TABLE IF NOT EXISTS dns_cache (
    host_or_triad TEXT PRIMARY KEY,
    resolved_endpoint TEXT NOT NULL,
    capabilities_json TEXT,         -- JSON payload from /.well-known/datamesh.json
    auth_type TEXT DEFAULT 'none',  -- 'keycloak', 'api_key', 'none'
    probe_status TEXT NOT NULL,     -- 'ACTIVE', 'UNREACHABLE', 'DEGRADED'
    latency_ms INTEGER,
    probed_at TEXT NOT NULL,
    ttl_seconds INTEGER DEFAULT 86400
);
CREATE INDEX IF NOT EXISTS idx_dns_status ON dns_cache(probe_status);

-- 3. Ephemeral Cryptographic Public Key Cache
CREATE TABLE IF NOT EXISTS keys_cache (
    identity_id TEXT NOT NULL,       -- e.g. 'github:andreschirinos'
    fingerprint TEXT NOT NULL,      -- SHA-256 fingerprint
    key_type TEXT NOT NULL,         -- 'ssh-ed25519', 'ssh-rsa'
    public_key TEXT NOT NULL,       -- Raw OpenSSH public key
    fetched_at TEXT NOT NULL,
    ttl_seconds INTEGER DEFAULT 604800, -- 7 days
    PRIMARY KEY(identity_id, fingerprint)
);

-- 4. Lineage Executions, Sources, Transformations, and Artifacts
CREATE TABLE IF NOT EXISTS executions (
    id TEXT PRIMARY KEY,             -- UUIDv4 or sha256 execution digest
    target_triad TEXT NOT NULL,
    script_path TEXT,
    entrypoint TEXT,
    command TEXT,
    status TEXT NOT NULL,            -- 'SUCCESS', 'FAILED', 'RUNNING'
    exit_code INTEGER,
    started_at TEXT NOT NULL,
    completed_at TEXT,
    duration_ms INTEGER,
    user_identity TEXT
);

CREATE TABLE IF NOT EXISTS lineage_sources (
    id TEXT PRIMARY KEY,             -- UUIDv4
    execution_id TEXT NOT NULL,
    source_name TEXT NOT NULL,
    source_uri TEXT NOT NULL,
    source_hash TEXT,
    format TEXT,
    cache_key TEXT,
    FOREIGN KEY(execution_id) REFERENCES executions(id) ON DELETE CASCADE,
    FOREIGN KEY(cache_key) REFERENCES cache_entries(key) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS lineage_transformations (
    id TEXT PRIMARY KEY,             -- UUIDv4
    execution_id TEXT NOT NULL,
    step_name TEXT NOT NULL,
    tool TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    query_text TEXT,
    parameters_json TEXT,
    FOREIGN KEY(execution_id) REFERENCES executions(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS artifacts (
    id TEXT PRIMARY KEY,             -- UUIDv4 or sha256 digest
    execution_id TEXT NOT NULL,
    path TEXT NOT NULL,
    sha256 TEXT NOT NULL,
    bytes_size INTEGER NOT NULL,
    format TEXT NOT NULL,
    cache_key TEXT,
    FOREIGN KEY(execution_id) REFERENCES executions(id) ON DELETE CASCADE,
    FOREIGN KEY(cache_key) REFERENCES cache_entries(key) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_exec_target ON executions(target_triad);
CREATE INDEX IF NOT EXISTS idx_sources_exec ON lineage_sources(execution_id);
CREATE INDEX IF NOT EXISTS idx_sources_cache ON lineage_sources(cache_key);
CREATE INDEX IF NOT EXISTS idx_trans_exec ON lineage_transformations(execution_id);
CREATE INDEX IF NOT EXISTS idx_art_exec ON artifacts(execution_id);
CREATE INDEX IF NOT EXISTS idx_art_cache ON artifacts(cache_key);
```

---

## 3. Mandatory Connection Invariants & Concurrency Rules

1. **Foreign Key Enforcement:** Runtimes MUST execute `PRAGMA foreign_keys = ON;` on every SQLite database connection to ensure cascading deletes and referential integrity between `executions` and `lineage_sources`.
2. **Concurrency (WAL Mode):** Runtimes MUST initialize SQLite connections with `PRAGMA journal_mode = WAL;`, `PRAGMA synchronous = NORMAL;`, and a busy timeout (`PRAGMA busy_timeout = 5000;`) to prevent database lock contention across concurrent CLI commands, background agent processes, and interactive notebooks.
3. **Migration Strategy:** Runtimes MUST maintain schema versioning using `PRAGMA user_version` and apply incremental, non-destructive migrations on startup.
4. **LRU Cache Eviction:** When local cached resources exceed configured capacity thresholds, runtimes evict files ordered by `last_accessed ASC`.
5. **Conditional HTTP 304 Revalidation:** For stale cache entries (`now > cached_at + ttl_seconds`), clients issue conditional HTTP requests with `If-None-Match: <etag>`. Upon receiving `304 Not Modified`, `last_accessed` is updated without re-downloading bytes.
