# Schema

Starlark schema files go here.

[API.md](API.md) is the reference for the DSL: every symbol, its signature, and what it does. It is generated from the docstrings in [kessel.star](kessel.star) by `make docs` and committed, so read it there rather than scrolling the source. For how the DSL works and what it emits, see [ARCHITECTURE.md](../ARCHITECTURE.md#starlark-dsl).

To invoke a KSL extension hand-authored downstream, use the `call_ksl_extension` builtin — no `load()` required. Wrap it in a function named for the extension rather than calling it directly from schema files; see [Backward-compatibility with KSL extensions](../ARCHITECTURE.md#backward-compatibility-with-ksl-extensions).

Those wrappers live in [extensions/](extensions), one file per KSL namespace that defines extensions (`rbac.star`, `hbi.star`). Calls to them are grouped one file per app at the top level (`advisor.star`, `patch.star`, …), mirroring the hand-authored `.ksl` files still in `rbac-config`. Every wrapper takes the reporter as its first argument, and the references it emits are written to that reporter's namespace — so `advisor.star` passes `"advisor"` and its references land in `advisor.json`. Keep the reporter matching the file name so each app's output stays in one place.