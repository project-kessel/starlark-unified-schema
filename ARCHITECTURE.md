# starlark-unified-schema Architecture

## Overview

starlark-unified-schema is a batch compiler that evaluates Starlark schema sources and produces build artifacts consumed by Kessel services. Schema authors define resources once under `schema/`; the Go interpreter emits two output families from the same source:

1. **JSON Schema** (Draft-07) — for inventory payload validation
2. **KSIL** (Kessel Schema Intermediate Language) — for authorization schema compilation

This is a transitional step toward a single canonical schema model. Services can onboard using the Starlark-based language while others continue with existing KSL and JSON Schema sources.

### Key Features

- **Single source, multiple output**: One Starlark definition drives both inventory validation and authorization schemas
- **Starlark DSL**: `schema/kessel.star` provides `resource()`, data types, relations, and permission expressions
- **Visitor-based codegen**: A pluggable `SchemaVisitor` interface keeps the processor output-agnostic
- **Cross-resource relations**: Metadata registry resolves relation targets across modules
- **Reporter namespaces**: Resources bind to Kessel reporters (`hbi`, `rbac`, `features`, …) that map to downstream layout

## Compilation Architecture

### Unified Stack

```
  Inputs                          Go interpreter                         Outputs              Downstream
  -----                          ----------------                        -------              ----------

  schema/**/*.star  ──┐
  schema/kessel.star  ├──▶  cmd/interpreter
                          │        │
                          │        ▼
                          │   lang.Loader  ──▶  execute .star, cache modules, record metadata
                          │        │
                          │        ▼
                          │   lang.Processor  ──▶  walk resource structs, dispatch to visitor
                          │        │
                          │        ▼
                          │   output.SchemaVisitor  (JSONSchemaVisitor or KSILVisitor)
                          │        │
                          │        ▼
                          └──   output.WriteSchemas
                                   │
                    ┌──────────────┴──────────────┐
                    ▼                             ▼
           JSON Schema (Draft-07)        KSIL JSON namespaces
                    │                             │
                    ▼                             ▼
             inventory-api              rbac-config → ksl → SpiceDB
```

**Key Design Decision**: Single processor with pluggable visitors, not separate compilers per output format.

- One semantic walk over Starlark `resource` values
- Each output format implements `SchemaVisitor` and ignores irrelevant callbacks (JSON Schema skips permissions; KSIL skips data fields)
- When both `JSONSCHEMA_OUTPUT_DIR` and `KSL_OUTPUT_DIR` are set, the CLI runs the full pipeline **once per visitor** — two independent passes with fresh visitor instances

## Output Formats

### 1. JSON Schema Output

**Consumer**: [inventory-api](https://github.com/project-kessel/inventory-api)

**Purpose**: Validates resource payloads submitted to the inventory API.

**Spec**: [JSON Schema Draft-07](http://json-schema.org/draft-07/schema#)

**Artifact layout**:

```
<JSONSCHEMA_OUTPUT_DIR>/
  <resource_type>/
    common_representation.json
    reporters/
      <reporter>/
        <resource_type>.json
```

Example:

```
output/jsonschema/host/common_representation.json
output/jsonschema/host/reporters/hbi/host.json
```

These map directly to inventory-api paths under `data/schema/resources/`.

**Behavior**:

| Input from processor | JSON Schema output |
|----------------------|-------------------|
| Data fields | Object properties with types, constraints, `required` |
| Relations | Mapped to data-field shapes by cardinality (`ExactlyOne` → required scalar; `Many` → array) |
| Permissions | Ignored |

Resources are grouped by **type name** (the Starlark variable name, e.g. `host`). Relations use the target resource's `id_type`, not a nested object schema.

### 2. KSIL Output

**Consumer**: [rbac-config](https://github.com/project-kessel/rbac-config) → `ksl` compiler → SpiceDB

**Purpose**: Authorization schema definitions derived from relations and permission expressions.

**Library**: [`github.com/project-kessel/ksl-schema-language`](https://github.com/project-kessel/ksl-schema-language) (`intermediate` types, `intermediate.Store()`)

**Artifact layout**:

```
<KSL_OUTPUT_DIR>/
  <reporter>.json
```

Example:

```
output/ksl/hbi.json
output/ksl/rbac.json
output/ksl/features.json
```

These are copied into rbac-config at `configs/<env>/schemas/src/`. The rbac-config `ksl` compiler accepts both text `.ksl` files and JSON KSIL `.json` files.

After updating KSIL in rbac-config, validate locally (do **not** overwrite committed `schema.zed` files):

```bash
cd /path/to/rbac-config
make init
make ksl-test-schema-stage   # writes to _private/test-schema/stage-schema.zed
```

**Behavior**:

| Input from processor | KSIL output |
|----------------------|-------------|
| Data fields | Ignored |
| Relations | `intermediate.Relation` with `self` body, target namespace/name, cardinality |
| Permissions | `intermediate.Relation` with expression body |

Resources are grouped by **reporter** (namespace), not by type name. One file per namespace contains all types defined for that reporter.

**Operator mapping**:

| Starlark / DSL | KSIL `RelationBody.Kind` |
|----------------|--------------------------|
| `intersect` / `and` | `intersect` |
| `union` / `or` | `union` |
| `exclude` / `unless` | `except` |
| `ref` | `reference` |
| `subref` | `nested_reference` |

Cardinality `Many` is converted to legacy `Any` for KSIL compatibility.

## Project Structure

```
starlark-unified-schema/
├── schema/                              # Starlark schema source (committed)
│   ├── kessel.star                      # Core DSL: resource(), field(), types, relations, permissions
│   ├── <type>/                          # One directory per logical resource type
│   │   ├── common_representation.star   # Shared fields/relations across reporters
│   │   └── reporters/<reporter>/        # Reporter-specific resource definitions
│   └── README.md
│
├── interpreter/                         # Go compiler
│   ├── cmd/interpreter/
│   │   └── main.go                      # CLI entry point, visitor wiring, output config
│   │
│   ├── internal/
│   │   ├── lang/                        # Starlark loading and semantic processing
│   │   │   ├── Loader.go                # Module execution, caching, metadata registry
│   │   │   ├── Processor.go             # Resource walk, visitor dispatch
│   │   │   ├── Builtins.go              # Predeclared builtins (struct, println)
│   │   │   └── Util.go                  # Starlark struct/dict helpers
│   │   │
│   │   ├── output/                      # Visitor implementations and writer
│   │   │   ├── Visitor.go               # SchemaVisitor interface and Members type
│   │   │   ├── jsonschema.go            # Draft-07 JSON Schema visitor
│   │   │   ├── ksil.go                  # KSIL namespace visitor
│   │   │   └── writer.go                # Filesystem output with path safety checks
│   │   │
│   │   └── util/
│   │       └── spy_visitor.go           # Test visitor with golden JSON assertions
│   │
│   ├── go.mod
│   └── README.md
│
├── references/                          # Example schemas (obsolete format; modeling only)
├── bin/                                 # Built interpreter binary (gitignored)
├── output/                              # Default generated artifacts (gitignored)
│   ├── jsonschema/
│   └── ksl/
├── .env                                 # Local output directory configuration
├── Makefile                             # build-interpreter, build-schema, test
├── ARCHITECTURE.md                      # This file
└── README.md
```

## Building and Running

```bash
# Build the interpreter
make build-interpreter

# Compile all schemas (uses .env for output directories)
make build-schema

# Run tests
make test
```

Alternatively, set environment variables manually:

```bash
export JSONSCHEMA_OUTPUT_DIR=output/jsonschema
export KSL_OUTPUT_DIR=output/ksl
./bin/interpreter
```

Compile specific files only:

```bash
./bin/interpreter schema/host/reporters/hbi/host.star
```

At least one of `JSONSCHEMA_OUTPUT_DIR` or `KSL_OUTPUT_DIR` must be set; the interpreter exits with an error if neither is configured.

## Core Concepts

### Compilation Flow

starlark-unified-schema uses a layered pipeline from Starlark source to disk artifacts:

```
1. Module Loading (lang.Loader)
   └─> Execute .star files via go.starlark.net
   └─> Resolve load() imports relative to schema root
   └─> Cache parsed module globals
   └─> Record metadata for each resource (reporter, type name, id_type)
        │
        ▼
2. Resource Discovery (lang.Processor)
   └─> Scan module globals for structs tagged kind="resource"
   └─> Skip non-resource values (common representation dicts, helpers)
        │
        ▼
3. Semantic Walk (lang.Processor → SchemaVisitor)
   └─> Visit common representation members (fields, relations, permissions)
   └─> Visit reporter-specific fields
   └─> Resolve cross-resource relation targets via metadata registry
   └─> Walk permission expression trees (and/or/unless/ref/subref)
        │
        ▼
4. Output Aggregation (SchemaVisitor.Results)
   └─> JSON Schema: group by type name → common + per-reporter schemas
   └─> KSIL: group by reporter namespace → one JSON file per namespace
        │
        ▼
5. Write (output.WriteSchemas)
   └─> Create directories, write files, validate paths stay within output root
```

### Starlark DSL

The DSL lives in `schema/kessel.star` and is plain Starlark — no custom interpreter builtins beyond `struct` and standard `load()`.

| Construct | Role |
|-----------|------|
| `resource(reporter, id_type, common={}, fields={}, permissions={})` | Defines a resource; returns a tagged struct consumed by the processor |
| `field(type=..., required=..., description=...)` | Data field member |
| `text`, `uuid`, `numeric_id`, `boolean`, `date_time`, `enum`, `nullable`, `union`, `array`, `object` | Data type constructors |
| `at_most_one`, `one`, `at_least_one`, `many`, `wildcard` | Relation cardinality helpers |
| `self()` | Relation target referring to the enclosing resource |
| `permissions={ "name": lambda proxy: ... }` | Permission factories evaluated at schema load time |

A minimal resource:

```python
load("kessel.star", "resource", "field", "uuid", "text", "nullable")
load("host/common_representation.star", common="host")

host = resource("hbi", common=common,
    id_type=uuid(),
    fields={
        "ansible_host": field(type=nullable(text(maxLength=255))),
    })
```

### Common vs Reporter Fields

Each resource type typically splits definition across two layers:

1. **`common_representation.star`** — Shared fields and relations imported by multiple reporters via `load(..., common="host")`. This is a plain dict, not a `resource()` call.
2. **`reporters/<reporter>/<type>.star`** — A `resource(reporter=..., ...)` binding that ties the type to a Kessel reporter namespace.

The processor merges common and reporter members when visiting a resource. JSON Schema emits them as separate files (`common_representation.json` plus per-reporter schemas).

### Permission Proxy Model

Permission expressions are evaluated in Starlark at schema load time, not in Go during the processor walk.

When `resource()` is called, it:

1. Builds a **proxy** object whose attributes mirror relation and permission names on the resource
2. Runs each permission factory (`lambda proxy: ...`) against that proxy
3. Stores the resulting permission body structs in the resource's `fields` dict

The proxy supports a fluent API: `intersect`, `union`, and `exclude` on reference nodes (`ref`, `subref`), enabling expressions like:

```python
permissions={
    "can_workspace_use_service": lambda s: s.does_workspace_have_service_preference.intersect(
        s.does_workspace_have_license
    ),
}
```

The Go processor only walks the resulting body trees; it does not interpret Starlark expressions.

### Metadata Registry and Load Order

Cross-resource relations (e.g. `one(workspace)`) require the target resource to be loaded first so its metadata is registered. The loader records `reporter`, type name, and `id_type` for every `resource` struct in a shared map keyed by struct identity.

Starlark `load()` imports enforce this ordering naturally — dependent modules are evaluated before modules that reference them.

### Artifact Promotion Boundary

Generated artifacts are gitignored in this repository. They are promoted to downstream repos via PRs:

- JSON Schema → inventory-api `data/schema/resources/`
- KSIL JSON → rbac-config `configs/<env>/schemas/src/`

Only Starlark source belongs in this repository.

## Component Interfaces

The compiler is built around well-defined interfaces that enable testability, flexibility, and extensibility.

### output.SchemaVisitor

**Purpose**: Receives semantic callbacks during the processor walk and accumulates output for serialization.

```go
type SchemaVisitor interface {
    BeginType(name string)
    VisitResource(typeName string, reporter string, commonMembers, reporterMembers *Members) error

    VisitDataField(name string, required bool, description *string, dataType any) any

    VisitTextDataType(minLength, maxLength *int, regex *string) any
    VisitUUIDDataType() any
    VisitNumericIDDataType(min, max *int) any
    VisitBooleanDataType() any
    VisitDateTimeDataType() any
    VisitEnumDataType(values []string) any
    VisitNullableDataType(inner any) any
    VisitCompositeDataType(dataTypes []any) any
    VisitArrayDataType(items any) any
    VisitObjectDataType(properties []any, required []string) any

    VisitAnd(left, right any) any
    VisitOr(left, right any) any
    VisitUnless(left, right any) any
    VisitReferenceExpression(name string) any
    VisitSubReferenceExpression(name, sub string) any

    VisitRelation(name, reporter, typeName, cardinality string, idType any) any

    BeginPermission(name string)
    VisitPermission(name string, body any) any

    Results() ([]OutputEntry, error)
}
```

**Key Types**:

```go
type Members struct {
    DataFields     []any
    RelationFields []any
    Permissions    []any
}

type OutputEntry struct {
    Path     string
    Contents []byte
}
```

**Implementations**:

- `JSONSchemaVisitor` — Draft-07 JSON Schema; ignores permission callbacks
- `KSILVisitor` — KSIL namespace JSON via `ksl-schema-language`; ignores data field callbacks
- `SpyVisitor` (test) — Captures visitor calls and can compare results against a JSON representation

### lang.Loader

**Purpose**: Executes Starlark modules and maintains the metadata registry.

```go
func NewLoader(path string) *Loader
func (l *Loader) Load(thread *starlark.Thread, name string) (starlark.StringDict, error)
func (l *Loader) GetAllModuleNames() ([]string, error)
func (l *Loader) SetMetadata(metadata map[resourceType]meta)
```

**Key Features**:

- Module caching after first execution
- Recursive discovery of `*.star` files under the schema root
- `Thread.Load` callback resolves `load()` imports
- Records `reporter`, type name, and `id_type` for each resource on load

**Builtins** (`Builtins.go`): `struct` (from starlarkstruct), `println` (debug)

### lang.Processor

**Purpose**: Walks Starlark `resource` values and dispatches to a `SchemaVisitor`.

```go
func NewProcessor(loader *Loader) *Processor
func (p *Processor) Process(visitor output.SchemaVisitor, files ...string) error
```

Member dispatch uses the `kind` attribute on each field struct:

| `kind` | Processor action |
|--------|------------------|
| `field` | Resolve data type recursively; `visitor.VisitDataField(...)` |
| `relation` | Resolve target resource via metadata; `visitor.VisitRelation(...)` |
| `permission` | Walk permission body tree; `visitor.VisitPermission(...)` |

### output.WriteSchemas

**Purpose**: Writes serialized output entries to disk with path traversal protection.

```go
func WriteSchemas(outputDir string, entries []OutputEntry) error
```

## Data Flow

### Component Interaction

```
CLI
 │
 ├─▶ Loader.NewLoader(schema/)
 ├─▶ Processor.NewProcessor(loader)
 └─▶ Visitor.NewJSONSchemaVisitor()  OR  Visitor.NewKSILVisitor()
      │
      │  [repeat for each configured output directory]
      │
      ├─▶ Processor.Process(visitor, files...)
      │    │
      │    │  [for each .star module]
      │    │
      │    ├─▶ Loader.Load(module)
      │    │    │
      │    │    ├─▶ Starlark.ExecFileOptions + load()
      │    │    │    └── returns globals (resource structs)
      │    │    │
      │    │    └─▶ Loader.recordMetadata(resources)
      │    │
      │    └─▶ [for each resource in module]
      │         Processor ──▶ Visitor.BeginType / VisitResource / ...
      │
      ├─▶ Visitor.Results()
      │    └── returns []OutputEntry
      │
      └─▶ WriteSchemas(outputDir, entries)
```

When both `JSONSCHEMA_OUTPUT_DIR` and `KSL_OUTPUT_DIR` are set, the outer loop runs twice — once per visitor — with independent visitor state.

### JSON Schema Compilation Flow

```
1. CLI reads JSONSCHEMA_OUTPUT_DIR
        │
        ▼
2. NewJSONSchemaVisitor()
        │
        ▼
3. Processor.Process(visitor, files...)
   For each .star module:
     Loader.Load() → execute Starlark, record metadata
     For each resource struct:
       Visit common members → data fields + relations (as schema shapes)
       Visit reporter fields → reporter-specific properties
        │
        ▼
4. JSONSchemaVisitor.Results()
   Group by type name
   Emit common_representation.json + reporters/<reporter>/<type>.json
        │
        ▼
5. WriteSchemas(JSONSCHEMA_OUTPUT_DIR, entries)
   Write Draft-07 JSON Schema files
```

### KSIL Compilation Flow

```
1. CLI reads KSL_OUTPUT_DIR
        │
        ▼
2. NewKSILVisitor()
        │
        ▼
3. Processor.Process(visitor, files...)
   For each .star module:
     Loader.Load() → execute Starlark, record metadata
     For each resource struct:
       Visit common members → relations + permissions
       Visit reporter fields → relations + permissions
        │
        ▼
4. KSILVisitor.Results()
   Group by reporter namespace
   Serialize via intermediate.Store() → <reporter>.json
        │
        ▼
5. WriteSchemas(KSL_OUTPUT_DIR, entries)
   Write KSIL namespace JSON files
```

## Key Design Patterns

### Visitor Pattern

All output generation flows through `SchemaVisitor`. The processor is output-agnostic; adding a third format means implementing the interface, registering it in `main.go` with an environment variable, and handling serialization in `Results()`.

Each visitor no-ops callbacks that do not apply:

- JSON Schema: permission expression methods return nil
- KSIL: data type methods return nil

### Metadata Registry

Cross-resource relation resolution uses a registry populated during module load:

```go
type meta struct {
    reporter string
    typeName string
    idType   *starlarkstruct.Struct
}
```

The processor looks up target resources by struct identity when visiting relation fields. This decouples relation resolution from output format.

### Dual-Pass CLI

Output formats are configured independently via environment variables. The CLI creates one visitor per configured directory and runs `Processor.Process` separately for each. This keeps visitors stateless and avoids mixing JSON Schema and KSIL aggregation in one pass.

### Dependency Injection

All components accept dependencies via constructors:

```go
loader := lang.NewLoader("schema")
processor := lang.NewProcessor(loader)

// JSON Schema pass
jsonVisitor := output.NewJSONSchemaVisitor()
if err := processor.Process(jsonVisitor); err != nil { /* ... */ }
entries, _ := jsonVisitor.Results()
output.WriteSchemas(os.Getenv("JSONSCHEMA_OUTPUT_DIR"), entries)

// KSIL pass
ksilVisitor := output.NewKSILVisitor()
if err := processor.Process(ksilVisitor); err != nil { /* ... */ }
entries, _ = ksilVisitor.Results()
output.WriteSchemas(os.Getenv("KSL_OUTPUT_DIR"), entries)
```

This enables:

- **Testability**: Swap real visitors with `SpyVisitor` or stubs
- **Flexibility**: Add output formats without modifying the processor
- **Clarity**: Explicit dependencies visible in `main.go`

## Testing Strategy

The visitor-driven design enables testing at multiple levels.

### Unit Tests

Each component can be tested in isolation with in-memory Starlark sources:

```go
reader := newInMemorySourceFileReader("schema")
processor := setupProcessorWithKessel(t, reader)

reader.AddFile("host/reporters/hbi/host.star", []byte(`...`))

spy := util.NewSpyVisitor()
err := processor.Process(spy)

spy.AssertJSON(t, `{
    "host": {
        "common": {"fields": [...]},
        "reporters": {"hbi": {"fields": [...]}}
    }
}`)
```

The `SpyVisitor` captures all visitor callbacks as a JSON tree and compares against golden expectations. Tests cover:

- Common + reporter field merging
- Data type constraints (text, uuid, enum, nullable, union, array, object)
- Relation cardinality and cross-resource references
- Permission expression trees (and, or, unless, ref, subref)

### Integration Tests

The CLI wiring is exercised indirectly through processor tests that use the real `kessel.star` DSL loaded from disk (`addRealSchemaFile`). Loader tests verify module discovery and `load()` resolution.

Run all tests:

```bash
make test
```

## Related Documentation

- **[README.md](README.md)**: Setup, usage, contributing workflow, and downstream promotion
- **[schema/README.md](schema/README.md)**: Schema directory overview
- **[interpreter/README.md](interpreter/README.md)**: Interpreter package overview
- **[references/](references/)**: Example schemas (obsolete format; modeling only)

### Related Repositories

| Repository | Role |
|------------|------|
| [starlark-unified-schema](https://github.com/project-kessel/starlark-unified-schema) | Starlark schema source and compiler (this repo) |
| [ksl-schema-language](https://github.com/project-kessel/ksl-schema-language) | KSL compiler and KSIL serialization library |
| [inventory-api](https://github.com/project-kessel/inventory-api) | Consumes generated JSON Schema |
| [rbac-config](https://github.com/project-kessel/rbac-config) | Consumes generated KSIL JSON |
