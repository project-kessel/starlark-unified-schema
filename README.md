# Starlark Unified Schema

A unified schema language and compiler for [Kessel](https://github.com/project-kessel). Schema authors write resource definitions once in [Starlark](https://github.com/bazelbuild/starlark/blob/master/spec.md), and the interpreter produces build artifacts consumed by Kessel services.

This is a transitional step toward a single schema model: services can onboard using the Starlark-based language while others continue to use existing KSL and JSON Schema sources. Over time, Starlark becomes the canonical definition.

## Overview

Kessel resources are defined with data fields (primitive types, other resource types), which services send to Kessel, and permissions (derived from graph traversal), which services can query from Kessel. Today those definitions are spread across multiple formats and repositories. This project centralizes authoring in Starlark and compiles to two output formats today:

| Output | Format | Consumed by | Purpose |
|--------|--------|-------------|---------|
| **JSON Schema** | Draft-07 JSON Schema files | [inventory-api](https://github.com/project-kessel/inventory-api) | Validates resource payloads submitted to the inventory API |
| **KSIL** (Kessel Schema Intermediate Language) | JSON namespace files | [rbac-config](https://github.com/project-kessel/rbac-config) | Input to the `ksl` compiler, which produces SpiceDB authorization schemas |

The Go interpreter loads `.star` files from the `schema/` directory, evaluates resource definitions (including cross-resource relations and permission expressions), and writes artifacts to configurable output directories.

## Related repositories

| Repository | Role |
|------------|------|
| [starlark-unified-schema](https://github.com/project-kessel/starlark-unified-schema) | **This repo** — Starlark schema source and compiler |
| [ksl-schema-language](https://github.com/project-kessel/ksl-schema-language) | KSL compiler and KSIL serialization library used by the interpreter |
| [inventory-api](https://github.com/project-kessel/inventory-api) | Inventory service; consumes generated JSON Schema under `data/schema/resources/` |
| [rbac-config](https://github.com/project-kessel/rbac-config) | RBAC and authorization config; consumes generated KSIL JSON under `configs/<env>/schemas/src/` |

## Repository layout

```
schema/                         # Starlark schema source files (.star)
  kessel.star                   # Core DSL: resource(), field(), relations, permissions
  <type>/                       # One directory per resource type (host, workspace, …)
    common_representation.star  # Shared fields/relations across reporters
    reporters/<reporter>/       # Reporter-specific resource definitions
interpreter/                    # Go compiler (Starlark → JSON Schema + KSIL)
  cmd/interpreter/              # CLI entry point
  internal/lang/                # Starlark loader and processor
  internal/output/              # JSON Schema and KSIL visitors
  internal/util                 # Additional helper types like the SpyVisitor used in tests
references/                     # Example schemas for reference (in an obsolete format / for modeling purposes only)
Makefile                        # build-interpreter, build-schema, test
.env                            # Output directory configuration (local, not committed with secrets)
output/                         # Generated artifacts (gitignored, primarily for debugging purposes)
  jsonschema/                   # JSON Schema output
  ksl/                          # KSIL JSON output
```

Each `.star` file that defines a `resource(...)` becomes input to the compiler. Files are discovered automatically from `schema/` unless specific paths are passed on the command line.

## Prerequisites

- [Go](https://go.dev/dl/) 1.25+ (see `interpreter/go.mod`)
- [Make](https://www.gnu.org/software/make/)
- [dotenv](https://github.com/theskumar/python-dotenv) CLI (optional but used: `dnf install python3-dotenv+cli` or `pip install python-dotenv[cli]`) — used by `make build-schema` to load `.env`

## Getting started

### 1. Clone the repository

```sh
git clone https://github.com/project-kessel/starlark-unified-schema.git
cd starlark-unified-schema
```

If you also plan to open PRs in the downstream repos, clone those alongside this project:

```sh
git clone https://github.com/project-kessel/inventory-api.git
git clone https://github.com/project-kessel/rbac-config.git
```

### 2. Configure output directories

Copy or edit `.env` to point at local output directories (the defaults write into `output/` inside this repo):

```sh
# JSON Schema artifacts → inventory-api layout
JSONSCHEMA_OUTPUT_DIR=output/jsonschema

# KSIL JSON artifacts → rbac-config layout
KSL_OUTPUT_DIR=output/ksl
```

To write directly into cloned downstream repos during development, point the variables at those paths instead:

```sh
JSONSCHEMA_OUTPUT_DIR=/path/to/inventory-api/data/schema/resources
KSL_OUTPUT_DIR=/path/to/rbac-config/configs/stage/schemas/src
```

At least one of `JSONSCHEMA_OUTPUT_DIR` or `KSL_OUTPUT_DIR` must be set; the interpreter exits with an error if neither is configured.

### 3. Build and run

```sh
# Build the interpreter binary to bin/interpreter (optional)
make build-interpreter

# Compile all schemas (builds interpreter, then runs it with .env)
make build-schema
```

Alternatively, set environment variables manually and invoke the interpreter directly (doesn't require dotenv):

```sh
make build-interpreter
export JSONSCHEMA_OUTPUT_DIR=output/jsonschema
export KSL_OUTPUT_DIR=output/ksl
./bin/interpreter
```

To compile specific files only:

```sh
./bin/interpreter schema/host/reporters/hbi/host.star
```

## Generated artifacts

### JSON Schema (for inventory-api)

For each resource type, the compiler writes:

```
<resource_type>/common_representation.json
<resource_type>/reporters/<reporter>/<resource_type>.json
```

Example output for the `host` resource:

```
output/jsonschema/host/common_representation.json
output/jsonschema/host/reporters/hbi/host.json
```

These map directly to [inventory-api](https://github.com/project-kessel/inventory-api) paths under `data/schema/resources/`.

### KSIL JSON (for rbac-config)

For each reporter namespace, the compiler writes one JSON file named after the namespace:

```
hbi.json
rbac.json
features.json
```

These are JSON-serialized [KSIL](https://github.com/project-kessel/ksl-schema-language) namespace definitions. These go to `configs/stage/schemas/src/` (or `configs/prod/schemas/src/`) in the rbac-config repository (see: step 2: Configure output directories.) The rbac-config `ksl` compiler accepts both text `.ksl` files and JSON KSIL `.json` files.

After updating KSIL files in rbac-config, validate the compiled schema locally (see rbac-config [Makefile](https://github.com/project-kessel/rbac-config/blob/master/Makefile) for 'test' targets, do **not** overwrite committed `schema.zed` files):

```sh
cd /path/to/rbac-config
make init
make ksl-test-schema-stage   # writes to _private/test-schema/stage-schema.zed
```

## Writing schemas

Schemas live under `schema/` and use the DSL defined in `schema/kessel.star`. A minimal resource ties together a reporter, identifier type, optional common representation, fields, and permissions:

```python
load("kessel.star", "resource", "field", "uuid", "text", "nullable")
load("host/common_representation.star", common="host")

host = resource("hbi", common=common,
    id_type=uuid(),
    fields={
        "ansible_host": field(type=nullable(text(maxLength=255))),
    })
```

Resources can reference other resources (relations) and define permission expressions using the logic operators (`intersect`, `union`, `exclude`) and helpers (`any`, `all`). See existing schemas under `schema/` for patterns.

## Contributing: end-to-end workflow

Schema changes typically touch up to three repositories. Work in this order:

### Step 1 — Update Starlark schemas (this repository)

1. Create a feature branch from `main`.
2. Edit or add `.star` files under `schema/`.
3. `make build-schema` to verify compilation.
4. Review generated output under `output/` (or your configured directories).
5. Open a PR in [starlark-unified-schema](https://github.com/project-kessel/starlark-unified-schema) with:
   - The Starlark source changes
   - A brief description of the schema change and which services/reporters are affected

Only Starlark source belongs in this repository; generated artifacts are not committed here.

### Step 2 — Update JSON Schema in inventory-api

1. Copy the relevant files from `output/jsonschema/` into [inventory-api](https://github.com/project-kessel/inventory-api) at `data/schema/resources/`, preserving the directory structure.
2. Open a PR in inventory-api with the updated JSON Schema files.
3. Follow inventory-api's contributing and CI requirements.

### Step 3 — Update KSIL in rbac-config

1. Copy the relevant namespace JSON files from `output/ksl/` into [rbac-config](https://github.com/project-kessel/rbac-config) at `configs/stage/schemas/src/` (and `configs/prod/schemas/src/` when promoting to production).
2. Run `make ksl-test-schema-stage` (or prod) in rbac-config to validate the compiled authorization schema.
3. Open a PR in rbac-config.

Coordinate PRs across repositories so schema, inventory validation, and authorization stay in sync. Link related PRs in each description when the change spans multiple repos.

## Development

### Debugging the interpreter

A VS Code launch configuration is provided in `.vscode/launch.json`. It runs `bin/interpreter` with output directories set to `output/jsonschema` and `output/ksl`. Build the debug binary first:

```sh
make build-interpreter-debug
```

### Makefile targets

| Target | Description |
|--------|-------------|
| `make build-interpreter` | Build the compiler to `bin/interpreter` |
| `make build-interpreter-debug` | Build with debug symbols for delve/VS Code |
| `make build-schema` | Build interpreter and compile all schemas |
| `make test` | Run Go unit tests |
| `make clean` | Remove `bin/` and `output/` |

## License

Apache License 2.0 — see [LICENSE](LICENSE).
