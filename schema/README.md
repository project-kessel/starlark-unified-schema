# Schema

Starlark schema files go here.

The DSL is [kessel.star](kessel.star). For how it works and what it emits, see [ARCHITECTURE.md](../ARCHITECTURE.md#starlark-dsl).

To invoke a KSL extension hand-authored downstream, use the `call_ksl_extension` builtin — no `load()` required. Wrap it in a function named for the extension rather than calling it directly from schema files; see [Backward-compatibility with KSL extensions](../ARCHITECTURE.md#backward-compatibility-with-ksl-extensions).

Those wrappers live in [extensions/](extensions), one file per KSL namespace that defines extensions (`rbac.star`, `hbi.star`). Calls to them are grouped one file per app at the top level (`advisor.star`, `patch.star`, …), mirroring the hand-authored `.ksl` files still in `rbac-config`. Every wrapper takes the reporter as its first argument, and the references it emits are written to that reporter's namespace — so `advisor.star` passes `"advisor"` and its references land in `advisor.json`. Keep the reporter matching the file name so each app's output stays in one place.