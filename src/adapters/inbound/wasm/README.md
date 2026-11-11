# WebAssembly Inbound Adapter (`src/adapters/inbound/wasm`)

## Overview & Role

Browser and JavaScript WebAssembly entrypoint (`//go:build js && wasm` in [`main_wasm.go`](./main_wasm.go)) wiring Open Knowledge Format (OKF) / DataMesh discovery, resolution, SQL querying, and validation to `globalThis.DataMesh`.

## Interface

Mounted via inline `js.FuncOf(...)` closures on `js.Global().Set("DataMesh", dataMeshObj)` inside `[main_wasm.go:main]`:
- **`DataMesh.discover(targetURL?: string): Promise<Catalog>`** (`[main_wasm.go:main]`)
- **`DataMesh.search(keyword: string, targetURL?: string): Promise<CatalogEntry[]>`** (`[main_wasm.go:main]`)
- **`DataMesh.resolve(uri: string): Promise<DataProduct>`** (aliased as **`DataMesh.get(uri: string)`**, `[main_wasm.go:main]`)
- **`DataMesh.query(resourceURI: string, optionsJSON?: string): Promise<QueryResult>`** (`[main_wasm.go:main]`)
- **`DataMesh.executeSQL(sqlQuery: string, optionsJSON?: string): Promise<QueryResult>`** (`[main_wasm.go:main]`)
- **`DataMesh.validate(target?: string): Promise<ValidationReport>`** (`[main_wasm.go:main]`)
- **`DataMesh.getConfig(): Config`** (synchronous, `[main_wasm.go:main]`)

## Invariants

- **Non-Blocking Promise Execution (`[main_wasm.go:main]`)**: Asynchronous operations spawn Go goroutines inside the JS `Promise` executor and marshal results via `js.Global().Get("JSON").Call("parse", ...)`.
- **Stateless Browser Storage (`[main_wasm.go:main]`)**: Passes `nil` for `StoragePort` (`usecases.NewResolveDataProductUseCase(nodeResolver, nil, cfg)` and `usecases.NewQueryDataProductUseCase(inmemEngine, nil, unifiedReader)`) with a `15s` HTTP timeout.
- **Event Loop Persistence (`[main_wasm.go:main]`)**: Blocks `main()` on an unclosed channel (`<-c`) to keep the Go WASM runtime alive.

## Side Effects

- **Global JS Namespace Mutation (`[main_wasm.go:main]`)**: Mutates `globalThis.DataMesh` in the host JavaScript environment and writes diagnostic messages via `fmt.Println` to the WASM console.

## Verification

```bash
cd src && GOOS=js GOARCH=wasm go build ./adapters/inbound/wasm
```
