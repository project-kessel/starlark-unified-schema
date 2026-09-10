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
| `internal/output` | `jsonschema_test.go` | JSONSchemaVisitor unit tests, per data type |
| `internal/output` | `ksil_test.go` | KSILVisitor unit tests, per relation/permission shape |
| `internal/util` | `spy_visitor_test.go` | SpyVisitor capture format |
| `cmd/interpreter` | `e2e_test.go` | Full pipeline: `.star` fixtures → JSON Schema and KSIL |

For pipeline and visitor design context, see [ARCHITECTURE.md](../ARCHITECTURE.md#testing-strategy).

## Processor tests (primary pattern)

End-to-end semantic behavior without writing files to disk:

1. Create an in-memory schema reader: `NewInMemorySourceFileReader("schema")`.
2. Load the real DSL: `setupProcessorWithKessel(t, reader)` — loads `kessel.star` from disk via `AddRealSchemaFile`.
3. Add inline `.star` fixtures with `util.AddFile(t, reader, ...)`.
4. Run `processor.Process(...)`.
5. Assert with `spy.AssertJSON(t, \`{...}\`)` — golden JSON of visitor callbacks.

Example skeleton:

```go
reader := NewInMemorySourceFileReader("schema")
processor := setupProcessorWithKessel(t, reader)

util.AddFile(t, reader, "host/common_representation.star", `...`)
util.AddFile(t, reader, "host/reporters/hbi/host.star", `...`)

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

`Loader_test.go` covers module discovery, `load()` dependency order, non-`.star` file filtering, and load caching. Use the same in-memory reader helpers (`createDefaultLoaderReaderAndThread`, `util.AddFile`).

## SpyVisitor

`internal/util/spy_visitor.go` implements `output.SchemaVisitor` and records callbacks as a JSON tree for golden comparison.

- Prefer extending `Processor_test.go` with new golden JSON when behavior changes.
- Add isolated SpyVisitor tests only when the visitor capture format itself changes.

## Output tests

Two layers cover what the visitors emit:

- `internal/output/jsonschema_test.go` and `ksil_test.go` drive visitor callbacks by hand (`v.BeginType(...)`, `v.VisitResource(...)`) with no reader or processor, then assert on `v.Results()`. Use these to test granular output behaviors, errors, etc.
- `cmd/interpreter/e2e_test.go` runs the real pipeline — `.star` fixtures through the processor into both visitors — use these for golden path tests through the entire pipeline.

`internal/util/test_helpers.go` holds the shared assertions:

- `util.VerifyKSILResults(t, visitor, map[string]*intermediate.Namespace{...})` — loads each emitted KSIL file, sorts types/relations, and compares to an expected namespace built in code.
- `util.VerifyJSONSchemaResults(t, visitor, map[string]util.JsonSchemaTestCase{...})` — validates `Valid` and `Invalid` sample documents against each emitted schema. Every emitted path needs an entry and vice versa, so a new output file fails the test until it's described.

Note that `internal/output`'s own tests are in `package output` and still carry private copies of these two helpers; they can't import `util` because `util` imports `output`. Prefer the `util` versions from any other package.

## Gaps and extension guidance

There are no golden output *files* checked into the repo — expectations live in Go code in the tests above. Schema-level output is still worth eyeballing via `make build-schema` after visitor changes.

## When to add tests

| Change | Action |
|--------|--------|
| Processor or Loader semantic behavior | Add or update SpyVisitor golden JSON in `Processor_test.go` |
| Loader discovery or import resolution | Extend `Loader_test.go` |
| How a data type or relation renders in an output format | Extend `internal/output/jsonschema_test.go` or `ksil_test.go` |
| DSL feature spanning the whole pipeline | Extend `cmd/interpreter/e2e_test.go` |
| SpyVisitor JSON shape | Update `spy_visitor_test.go` and existing golden expectations |
| Starlark-only schema change with no DSL change | `make build-schema` is usually sufficient; add processor tests if the change exercises new semantic paths |
