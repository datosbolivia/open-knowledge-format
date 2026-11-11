# RFC 0001: OKF Peer-to-Peer (P2P) Sovereign Network Specification (OKF-P2P)

* **Status:** Experimental / Non-Normative Extension
* **Date:** 2026-11-11
* **Target Profile:** `level-2-sovereign`

---

## 1. Motivation & Scope

To eliminate single points of failure, resist censorship, and guarantee data sovereignty across federated institutions without centralized cloud providers, this RFC defines an optional **Peer-to-Peer (P2P) and Distributed Network Protocol** for the Open Knowledge Format (OKF-P2P).

A standard OKF v0.2 bundle does not require P2P networking; static HTTP/S, S3/MinIO, and Git repositories satisfy core requirements. Nodes that explicitly advertise OKF-P2P conformance implement the specifications described herein.

---

## 2. Node Identity & Content Addressing

1. **Sovereign Node Identity (`NodeID`):**
   - Each sovereign node in the OKF network possesses an Ed25519 cryptographic key pair.
   - The canonical `NodeID` is derived as the multihash digest of the node's public key (compatible with libp2p PeerIDs, e.g., `12D3KooW...`).
   - Nodes sign all catalog announcements, metadata manifests, and resolution answers with this private key.

2. **Immutable Content Addressing (CID):**
   - Every physical resource (static Parquet, CSV, or manifest) is addressed by its cryptographic content identifier (multihash SHA-256 or BLAKE3).
   - The data address depends strictly on its bit content, ensuring zero-tampering guarantees across distributed peers.

---

## 3. Decentralized Naming & Resolution Protocol ("Sovereign Data DNS")

ODKF defines a decentralized naming and query protocol analogous to the Internet Domain Name System (DNS), operated over a Distributed Hash Table (DHT) and cryptographic registries.

### 3.1 Resolution Syntax & Grammar

Clients support reverse-lookup query syntax:
```text
<dataset>.<resource>@<catalog>
```
Or fully-qualified sovereign URI notation:
```text
okf://<catalog>/<dataset>/<resource>
```

### 3.2 The `okf-dig` Diagnostic and Resolution Command

A conforming node or client tool implements a resolution utility:
```bash
# Querying a sovereign dataset resource
okf-dig elecciones_2025.votos_municipales@bolivia

# Diagnostic query with full cryptographic lineage and peer records
okf-dig +trace +crypto calidad_aire.pm25_horario@lapaz_alcaldia
```

### 3.3 OKF Sovereign Resource Record Types

Resolution queries return typed cryptographic records:

| Record Type | Name | Purpose | Example Payload |
| :--- | :--- | :--- | :--- |
| **`SOA`** | Start of Authority | Identifies the sovereign owner node, public key, trust tier, and contact multiaddr. | `NodeID: 12D3... PubKey: ed25519:... Tier: tier_1_verified Serial: 2026102101` |
| **`CAT`** | Catalog Manifest | Points to the CID of the catalog manifest (`llms.txt`, `index.md`, or `catalog.yaml`) and version. | `CID: QmXoy... Version: 0.2.0 Status: active` |
| **`RES`** | Resource Location | Physical locations where the dataset/resource can be fetched (multiaddresses, HTTP, S3, P2P Bitswap). | `Format: parquet Hash: sha256:e3b0c4... Multiaddrs: [/ip4/190.181.2.10/tcp/4001/p2p/12D3..., https://datos.gob.bo/data.parquet]` |
| **`SCH`** | Syntactic Schema | Direct link to the Frictionless / JSON Schema contract governing the columns. | `Type: frictionless_table_schema Hash: sha256:a1b2... Path: schemas/pm25.json` |
| **`LIN`** | Lineage & PROV-O | Activity URI, transformation script hash, and upstream inputs (`prov:wasGeneratedBy`). | `Activity: urn:uuid:... ScriptHash: sha256:9f8e... Sources: [senamhi:meteo:raw]` |
| **`SRV`** | Dynamic Service | Connection endpoint for dynamic / streaming queries (DuckDB Flight SQL, Arrow Flight, WebSocket). | `Proto: flight-sql Port: 8443 TLS: true Auth: ssh-token` |

### 3.4 Resolution Process via Kademlia DHT

1. **Routing lookup:** The resolving client maps the target catalog name (`<catalog>`) to its 256-bit DHT key (`SHA-256("okf:catalog:" + catalog)`).
2. **Peer retrieval:** The client queries the nearest $k$-peers in the DHT to obtain provider records of the sovereign node managing `<catalog>`.
3. **Record verification:** The sovereign node returns signed `SOA` and `CAT` records. The client verifies the Ed25519 signature against trusted public keys or federated trust lists.
4. **Sub-resource resolution:** Using the catalog's CID, the client retrieves the concept document for `<dataset>` and extracts the `RES` record for `<resource>`.
5. **Transport selection:** The client chooses the fastest or lowest-cost transport available (local LAN peer via mDNS > direct TCP/QUIC peer > remote HTTPS fallback).

---

## 4. Connectivity, Multiaddresses and Transport Protocols

1. **Multiaddress Standard:**
   Nodes and data endpoints MUST be addressable via the `multiaddr` standard:
   ```text
   /ip4/200.87.120.45/tcp/4001/p2p/12D3KooWDpJ7As7BWAbr.../okf/catalog/lapaz/dataset/aire
   /ip6/2800:cd0:.../udp/4001/quic-v1/p2p/12D3KooW.../okf/stream/flight
   ```

2. **Transport Agnosticism:**
   - **QUIC / UDP:** Default WAN transport for low-latency multiplexed streams and hole-punching.
   - **TCP with TLS 1.3:** Enterprise fallback for networks with strict UDP filtering.
   - **WebSockets / WebRTC:** Browser-based clients, edge dashboards, and WASM environments (e.g., DuckDB-WASM in client browsers).

3. **NAT Traversal & Hole Punching:**
   - Nodes behind residential or university NATs implement ICE (Interactive Connectivity Establishment), STUN, and AutoNAT.
   - When direct connection fails, Circuit Relay (Relay v2) allows mediated transport through authorized hub nodes without compromising end-to-end payload encryption.

---

## 5. GossipSub & Federated Cache Synchronization

1. **Publish-Subscribe Topics:**
   - `/okf/announce/<catalog>`: Announcements of new bundles, version bumps, and log updates.
   - `/okf/revocation`: Real-time cryptographic revocation of compromised keys or corrupt datasets.
   - `/okf/heartbeat`: Liveness and bandwidth capacity telemetry.

2. **Peer Discovery:**
   - **Local Discovery:** Zero-configuration discovery via mDNS within institutional networks.
   - **Global Discovery:** Kademlia DHT bootstrap list maintained by consortium nodes.

---

## 6. Security Analysis & Sybil Defenses

1. **Distinction Between Integrity and Sybil Resistance:**
   - Content hashing (SHA-256 / BLAKE3 CIDs) guarantees **immutable payload integrity** after a resource is retrieved.
   - However, content hashing alone **does not protect DHT routing tables against Sybil or Eclipse attacks**. To prevent malicious nodes from monopolizing neighborhood keys near target catalogs, participating DHT bootstrap networks MUST employ authenticated node identity tiers (e.g., stake-weighted peer scoring, verifiable credential registries, or federated consortium keylists).
2. **P2P Zero-Microdata Policy:**
   - When a concept specifies `policy: zero_microdata`, conforming peer nodes MUST refuse binary block transfers of raw microdata to unauthenticated peers, offering instead dynamic SQL aggregate queries via Flight SQL over authenticated TLS sessions.
