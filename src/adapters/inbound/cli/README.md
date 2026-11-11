# CLI & HTTP/MCP Server Adapter (`src/adapters/inbound/cli`)

## Overview & Role

Composition root for the `datamesh` CLI binary ([`main.go`](./main.go)), multi-protocol HTTP REST / CORS proxy / Model Context Protocol (MCP) JSON-RPC 2.0 server ([`serve.go`](./serve.go)), and network/filesystem boundary guards ([`security.go`](./security.go)).

## Interface

- **CLI Entrypoint (`[main.go:main]`)**: Dispatches subcommands `catalog [url]`, `search <keyword> [url]`, `get <uri_or_url>`, `sql "<SELECT ...>"`, `query <resource_uri> [--filter k=v] [--limit N]`, `validate <target>` (exits `2` when `!report.Valid`), `lineage [target] [--format|-f ascii|mermaid|json]`, `config`, and `serve [--port|-p 8000] [--host 127.0.0.1]`.
- **HTTP & MCP Server (`[serve.go:ServerConfig]`, `[serve.go:StartDataMeshServer]`)**:
  - `GET /health`: Service status and version (`0.3.0-dev.1`).
  - `GET /.well-known/datamesh.json`, `GET /.well-known/okf.json`: Sovereign node discovery metadata (`schema_version`, `catalog`, `auth`, `capabilities`).
  - `GET /proxy?url=<target_url>`: SSRF-guarded CORS proxy (`30s` timeout).
  - `GET /api/catalog?url=<optional_url>`: Federated or targeted catalog discovery.
  - `GET|POST /api/sql`: Executes SQL subset over triads (`?q=` or `{"query": "..."}`).
  - `POST /mcp`: JSON-RPC 2.0 MCP endpoint (`initialize`, `tools/list`, `tools/call`, `ping`) returning HTTP `202 Accepted` on JSON-RPC notifications (missing `id`, `id: null`, `notifications/*`, or `initialized`) and exposing tools `datamesh_discover_catalogs`, `datamesh_search_catalog`, `datamesh_get_dataproduct`, `datamesh_query_resource`, `datamesh_sql_query` (alias `query_sql`), and `read_resource`.
- **Security Guards ([`security.go`](./security.go))**:
  - `[security.go:hostIsBlocked]`: `hostIsBlocked(host string) bool`
  - `[security.go:validateOutboundURL]`: `validateOutboundURL(raw string) (string, error)`
  - `[security.go:isRemoteOrTriad]`: `isRemoteOrTriad(ref string) bool`
  - `[security.go:isForbiddenLocalRef]`: `isForbiddenLocalRef(ref string) bool`

## Invariants

- **SSRF Guard (`[security.go:hostIsBlocked]`, `[security.go:validateOutboundURL]`)**: Resolves DNS (`net.LookupIP`) and blocks `localhost`, empty hosts, loopback, RFC1918 private, link-local unicast/multicast, multicast, and unspecified addresses; permits only `http` and `https` schemes.
- **Local File Rejection (`[security.go:isForbiddenLocalRef]`, `[security.go:isRemoteOrTriad]`)**: Rejects `file://`, UNC `\\`, absolute paths, Windows drive-letter paths (`C:\...`), and `..` traversal segments on remote MCP resource queries.
- **Server Timeouts (`[serve.go:StartDataMeshServer]`)**: Enforces `ReadHeaderTimeout: 10s`, `ReadTimeout: 30s`, `WriteTimeout: 60s`, `IdleTimeout: 120s`.

## Side Effects

- **Process, Network & Filesystem I/O**: Binds TCP listener in `[serve.go:StartDataMeshServer]`, performs DNS lookups (`net.LookupIP`) and outbound HTTP requests, initializes local cache directory (`FileStorage`), writes to `os.Stdout`/`os.Stderr`, and terminates via `os.Exit` on fatal CLI errors.

## Verification

```bash
cd src && go test ./adapters/inbound/cli/...
```
