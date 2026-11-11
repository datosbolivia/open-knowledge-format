# Configuration Loader Adapter (`src/adapters/outbound/config`)

## Overview & Role

Zero-dependency configuration loader (`FileAndEnvConfigLoader` in [`loader.go`](./loader.go)) implementing [`outbound.ConfigLoaderPort`](../../../ports/outbound/README.md) with layered precedence: Environment Variables > Config File (`datamesh.json`, `datamesh.yaml`, `datamesh.yml`) > Domain Defaults.

## Interface

- **`[loader.go:FileAndEnvConfigLoader]`**: `type FileAndEnvConfigLoader struct`
- **`[loader.go:NewConfigLoader]`**: `NewConfigLoader() *FileAndEnvConfigLoader` — Initializes loader using `os.Environ`.
- **`[loader.go:FileAndEnvConfigLoader.Load]`**: `(l *FileAndEnvConfigLoader) Load(explicitPath string) (domain.Config, error)`:
  1. Initializes `domain.NewDefaultConfig()` (`https://datosbolivia.github.io/llms.txt`, `30s` timeout).
  2. Auto-discovers `datamesh.json`, `datamesh.yaml`, or `datamesh.yml` when `explicitPath == ""`, or reads `explicitPath`.
  3. Overlays `DATAMESH_*` base environment variables (`DATAMESH_CATALOG_URL`, `DATAMESH_CATALOGS`, `DATAMESH_CATALOG_URLS`, `DATAMESH_STORAGE_PATH`, `DATAMESH_TIMEOUT`, `DATAMESH_LOG_LEVEL`) and `DATAMESH__{ADAPTER}_{KEY}` / `DATAMESH__{ADAPTER}__{KEY}` adapter-scoped credentials.

## Invariants

- **Case Normalization (`[loader.go:FileAndEnvConfigLoader.Load]`)**: Adapter names and keys from JSON, YAML, and `DATAMESH__` environment variables are normalized to lowercase (`strings.ToLower`).
- **Missing File Tolerance (`[loader.go:FileAndEnvConfigLoader.Load]`)**: Missing implicit or explicit config files (`os.IsNotExist`) do not fail `Load`; malformed JSON returns an unmarshal error.

## Side Effects

- **Environment & Filesystem Reads (`[loader.go:NewConfigLoader]`, `[loader.go:FileAndEnvConfigLoader.Load]`)**: Reads process environment variables (`os.Environ`) and local configuration files (`os.Stat`, `os.ReadFile`).

## Verification

```bash
cd src && go test ./adapters/outbound/config/...
```
