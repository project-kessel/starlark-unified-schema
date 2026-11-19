# Schema

Starlark schema files go here.

[API.md](API.md) is the reference for the DSL: every symbol, its signature, and what it does. It is generated from the docstrings in [kessel.star](kessel.star) by `make docs` and committed, so read it there rather than scrolling the source. For how the DSL works and what it emits, see [ARCHITECTURE.md](../ARCHITECTURE.md#starlark-dsl).

To invoke a KSL extension hand-authored downstream, use the `call_ksl_extension` builtin — no `load()` required. Wrap it in a function named for the extension rather than calling it directly from schema files; see [Backward-compatibility with KSL extensions](../ARCHITECTURE.md#backward-compatibility-with-ksl-extensions).

Those wrappers live in [extensions/](extensions), one file per KSL namespace that defines extensions (`rbac.star`, `hbi.star`). Calls to them are grouped one file per app at the top level (`advisor.star`, `patch.star`, …), mirroring the hand-authored `.ksl` files still in `rbac-config`. Set the default once per file with `reporter("advisor")`, then omit the reporter from wrapper calls. The references land in that reporter's namespace (`advisor.json` in this example). An explicit `reporter="patch"` keyword overrides the default for one call. Defaults are scoped to the file being evaluated and do not propagate through `load()`.

```python
load("extensions/rbac.star", "add_v1_based_permission")

reporter("advisor")
add_v1_based_permission("advisor", "weekly_report", "read", "advisor_weekly_report_view")
add_v1_based_permission(
    reporter="patch",
    app="patch",
    resource="systems",
    verb="read",
    v2_perm="patch_system_view",
)
```
