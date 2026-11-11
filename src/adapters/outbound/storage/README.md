# Local Filesystem Storage Adapter (`src/adapters/outbound/storage`)

## Overview & Role

Thread-safe local filesystem adapter (`FileStorage` in [`file_storage.go`](./file_storage.go)) implementing [`outbound.StoragePort`](../../../ports/outbound/README.md) with strict directory confinement and atomic file writes.

## Interface

- **`[file_storage.go:FileStorage]`**: `type FileStorage struct`
- **`[file_storage.go:NewFileStorage]`**: `NewFileStorage(baseDir string) (*FileStorage, error)` — Cleans `baseDir` and ensures it exists (`os.MkdirAll(cleanBase, 0755)`).
- **`[file_storage.go:FileStorage.Save]`**: `(s *FileStorage) Save(ctx context.Context, key string, data []byte) error` — Writes to `<target>.tmp` (`0644`) and atomically renames to `<target>`.
- **`[file_storage.go:FileStorage.Load]`**: `(s *FileStorage) Load(ctx context.Context, key string) ([]byte, error)` — Reads bytes from `<baseDir>/<key>`.
- **`[file_storage.go:FileStorage.Exists]`**: `(s *FileStorage) Exists(ctx context.Context, key string) (bool, error)` — Reports whether `<baseDir>/<key>` exists on disk.

## Invariants

- **Path Traversal Prevention (`[file_storage.go:FileStorage.resolveSafePath]`)**: Cleans `key`, rejects any key containing `".."`, and verifies via `filepath.Rel(s.baseDir, fullPath)` that the resolved path never escapes `baseDir`.
- **Concurrency Safety (`[file_storage.go:FileStorage]`)**: Guards `Save` with `s.mu.Lock()` and `Load`/`Exists` with `s.mu.RLock()`.

## Side Effects

- **Filesystem Mutations (`[file_storage.go:NewFileStorage]`, `[file_storage.go:FileStorage.Save]`)**: Creates directories (`os.MkdirAll`), writes temporary `.tmp` files (`os.WriteFile`), and performs atomic renames (`os.Rename`) inside `baseDir`.

## Verification

```bash
cd src && go test ./adapters/outbound/storage/...
```
