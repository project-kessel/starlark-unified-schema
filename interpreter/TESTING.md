# Interpreter testing

Go tests for the compiler live under `interpreter/`. Run from the repository root:

```bash
make test          # go test -C ./interpreter/ -count=1 ./...
make build-schema  # after .star or visitor/output changes; inspect output/
```

For authorization-heavy schema changes, validate KSIL in a local [rbac-config](https://github.com/project-kessel/rbac-config) clone per [README.md](../README.md); do not overwrite committed `schema.zed` files.

Tests use the standard library `testing` package and `github.com/stretchr/testify`.

## Layout

| Package | File | Focus |
|---------|------|-------|
| `internal/lang` | `Processor_test.go` | Semantic walk: fields, relations, permissions |
| `internal/lang` | `Loader_test.go` | Module discovery, `load()` order, caching |
| `internal/util` | `spy_visitor_test.go` | SpyVisitor capture format |

For pipeline and visitor design context, see [ARCHITECTURE.md](../ARCHITECTURE.md#testing-strategy).

## Processor tests (primary pattern)

End-to-end semantic behavior without writing files to disk:

1. Create an in-memory schema reader: `newInMemorySourceFileReader("schema")`.
2. Load the real DSL: `setupProcessorWithKessel(t, reader)` — loads `kessel.star` from disk via `addRealSchemaFile`.
3. Add inline `.star` fixtures with `reader.AddFile(...)`.
4. Run `processor.Process(...)`.
5. Assert with `spy.AssertJSON(t, \`{...}\`)` — golden JSON of visitor callbacks.

Example skeleton:

```go
reader := newInMemorySourceFileReader("schema")
processor := setupProcessorWithKessel(t, reader)

reader.AddFile("host/common_representation.star", []byte(`...`))
reader.AddFile("host/reporters/hbi/host.star", []byte(`...`))

spy := util.NewSpyVisitor()
if err := processor.Process(spy); err != nil {
    t.Fatalf("Process failed: %v", err)
}

spy.AssertJSON(t, `{ ... }`)
```

Use this pattern when adding or changing:

- Common + reporter field merging
- Data types and constraints
- Relation cardinality and cross-resource references
- Permission expression trees (`intersect`/`union`/`exclude`, `ref`/`subref`, `any`/`all`)

For cross-resource relations, define or load dependency modules before modules that reference them — same ordering constraints as production `load()` imports.

## Loader tests

`Loader_test.go` covers module discovery, `load()` dependency order, non-`.star` file filtering, and load caching. Use the same in-memory reader helpers (`createDefaultLoaderReaderAndThread`, `reader.AddFile`).

## SpyVisitor

`internal/util/spy_visitor.go` implements `output.SchemaVisitor` and records callbacks as a JSON tree for golden comparison.

- Prefer extending `Processor_test.go` with new golden JSON when behavior changes.
- Add isolated SpyVisitor tests only when the visitor capture format itself changes.

## Gaps and extension guidance

There are no JSON Schema or KSIL golden-file integration tests in-repo today. Output correctness is validated indirectly via SpyVisitor and manual `make build-schema` inspection. Update these instructions to capture the patterns for those tests once they're established.

## When to add tests

| Change | Action |
|--------|--------|
| Processor, Loader, or visitor semantic behavior | Add or update SpyVisitor golden JSON in `Processor_test.go` |
| Loader discovery or import resolution | Extend `Loader_test.go` |
| SpyVisitor JSON shape | Update `spy_visitor_test.go` and existing golden expectations |
| Starlark-only schema change with no DSL change | `make build-schema` is usually sufficient; add processor tests if the change exercises new semantic paths |
