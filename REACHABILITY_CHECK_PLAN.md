# Plan: CheckRequest Reachability Verifier

**Status:** ready to implement
**Audience:** the agent implementing this feature. Read `GRAPH.md` first —
especially the **"Check cost"** section — then this document.

> Clean-room note: implement this fresh. Do **not** consult or reuse the
> `-paths` / `Reach` / `kesselReachPaths` "path explorer" work that exists on
> `origin/schema-playground` (commit `4e104f6`) — the maintainer has chosen to
> build this independently. You **do** build on the *cost* analysis
> (`analyze.ExplainCheck`, commit `0b57acd`) that is already in the local tree;
> that is a separate, in-tree feature. Pick names that will not collide with the
> remote work if the branches are ever merged (this plan uses `CheckReachable`,
> `-reach`, `kesselCheckReach` — distinct from `Reach`/`-paths`/`kesselReachPaths`).

## Summary

Add a **structural reachability verifier** modelled on inventory-api's
[`CheckRequest`](https://buf.build/project-kessel/inventory-api/docs/main:kessel.inventory.v1beta2#kessel.inventory.v1beta2.CheckRequest).
Given three inputs —

- **object** — a resource-type reporter facet (`TYPE.REPORTER`)
- **relation** — one permission-or-relation name on that object
- **subject** — a resource-type reporter facet (`TYPE.REPORTER`)

— it answers: **does the schema wiring needed to satisfy `object#relation@subject`
exist?** It walks the permission rewrite from `object#relation` down the relation
topology and reports whether at least one path terminates at a relation whose
target is `subject`, then **highlights that path (or paths) on the graph**.

It uses **no instance data and no IDs** — this is a purely static property of the
schema graph + rewrite trees. We are not replicating the resolver's evaluation;
we are verifying that the structure required to complete the query successfully is
present. **"No path exists" is a first-class answer** — that is how a schema
author learns the schema does not support an intended check.

## Design principles (inherited from the codebase — do not violate)

1. **Schema-native, never target-native.** No SpiceDB / KSIL / JSON-Schema
   concepts. Work only in the schema's own vocabulary.
2. **One artifact, many consumers.** The analysis is one Go implementation in
   `internal/analyze`, consumed by the CLI (`cmd/graph-analyze`) and the browser
   (`cmd/graph-wasm` → playground), each reading only the documented `graph.json`
   contract via `internal/graphdoc`.
3. **Deterministic + golden-tested.** Sort everything by a stable key. CLI and
   WASM outputs must be byte-identical, pinned by a parity golden test (mirror
   `TestReachMatchesFilesystem` naming from the check parity tests).
4. **Build on `ExplainCheck`, don't fork it.** `analyze.ExplainCheck` already
   walks the rewrite and returns a `*CheckNode` proof tree with the recursion
   guard and cross-facet name resolution. Reachability is a **pure post-walk** of
   that tree. Do not write a second resolver.

## Semantics (nail these before coding)

A check `object#relation@subject` is **structurally reachable** iff, expanding
`relation`'s rewrite from `object`, some path through
`reference` / `subreference` / `and` / `or` / `unless` ends at a **relation leaf
whose `(TargetType, TargetReporter)` equals `subject`**.

Decisions baked in (confirmed with the maintainer):

- **Subject granularity = full facet.** A leaf matches only when **both** the
  target type **and** the target reporter equal the subject facet. (Relation-edge
  targets always carry a resolved `TargetReporter` in `graph.json`, so even
  subject-only types like `user` appear as `user.<reporter>` facets.)
- **Traverse all branches; tag exclusions.** Walk both operands of `and`, `or`,
  **and `unless`** — we are verifying that wiring *exists*, not evaluating truth.
  A witness that reaches `subject` only by descending into the **right operand of
  an `unless`** (the exclusion/deny side) is recorded but **flagged as an
  exclusion path**. Rationale: the wiring is present, but reaching the subject
  there means it would be *denied*, not granted — the author must see that
  distinction.
- **Highlight the union of all witnesses.** Every hop on every witness path is
  highlighted; hops that appear **only** on exclusion paths get a distinct
  exclusion style. The rest of the graph is dimmed.

Verdict tiers derived from the witnesses:

| Verdict          | Condition                                                      |
| ---------------- | ------------------------------------------------------------- |
| `reachable`      | ≥1 witness path with `Excluded == false` (a real grant path) |
| `exclusion-only` | no grant witness, but ≥1 witness with `Excluded == true`      |
| `unreachable`    | no witness path at all                                        |

### Recursion is sound to treat as a dead-end for witnesses

`ExplainCheck` cuts a rewrite cycle with a `recursive` sentinel node keyed by
`facet#name`. A `recursive` node re-enters a permission **already expanded higher
on the path**, so it introduces **no new subject facet** — every subject facet
reachable through the cycle was already discovered before the cut (the guard only
fires on a true `facet#name` repeat; distinct facets in a hierarchy are expanded
until they repeat). Therefore, for witness extraction, a `recursive` node emits
**no terminal**. State this in a comment; do not attempt to unroll cycles.

## The core walk (build on the proof tree)

### Step 0 — additive change to `ExplainCheck` relation leaves (`check.go`)

Today only `arrow` nodes carry `TargetType`/`TargetReporter`. Relation **leaves**
must also carry them, so a leaf names its reachable subject facet. In
`resolveOn`, when `name` resolves to a relation edge `e`:

```go
n := &CheckNode{
    Kind: "relation", TypeName: f.TypeName, Reporter: f.Reporter,
    Name: name, Cardinality: e.Cardinality,
    TargetType: e.Target, TargetReporter: e.TargetReporter, // NEW (additive JSON)
}
```

The `CheckNode.TargetType`/`TargetReporter` fields already exist. This is purely
additive. **Update the `-check` golden** (`internal/analyze/check_test.go` and any
web parity golden in `internal/web`) — cost/behavior is unchanged, only new leaf
fields appear in the JSON.

### Step 1 — witness extraction (new file `internal/analyze/reach.go`)

Pure functions over the `*CheckNode` tree from `ExplainCheck`. No graph re-walk.

```go
// WitnessHop is one relation traversal on a path from object to subject.
type WitnessHop struct {
    FromType, FromReporter string
    Relation, Cardinality  string
    ToType, ToReporter     string
    Fanout                 bool // many-cardinality relation
}

// WitnessPath is one structural path object#relation ... -> subject.
type WitnessPath struct {
    Hops     []WitnessHop
    Excluded bool // path descends through the right operand of an `unless`
}

// ReachVerdict is the result of a CheckRequest reachability verification.
type ReachVerdict struct {
    Object   FacetRef      // TYPE.REPORTER
    Relation string
    Subject  FacetRef      // TYPE.REPORTER (required)
    Verdict  string        // "reachable" | "exclusion-only" | "unreachable"
    Paths    []WitnessPath // all witnesses (grant + exclusion), stable-sorted
    Proof    *CheckNode    // the underlying ExplainCheck tree (for a tree view)
}

// CheckReachable runs ExplainCheck, then extracts every witness path whose
// terminal relation leaf targets `subject` (matched on BOTH type and reporter).
func CheckReachable(doc graphdoc.Document, object FacetRef, relation string, subject FacetRef) (*ReachVerdict, error)
```

**Extraction DFS** over the proof tree, carrying `hops []WitnessHop` and a
`underUnlessRight bool`:

- `permission` → recurse into `Body` (name is cosmetic).
- `relation` (leaf) → if `(TargetType,TargetReporter) == subject`, emit a
  `WitnessPath{Hops: hops + thisLeafAsHop, Excluded: underUnlessRight}`. Else
  nothing (dead-end).
- `op` `and`/`or` → recurse into both children with the same `underUnlessRight`.
- `op` `unless` → recurse left with `underUnlessRight` unchanged; recurse right
  with `underUnlessRight = true`.
- `arrow` (subreference) → append a `WitnessHop` for this arrow's relation, then
  recurse into `Children[0]` (the target sub-evaluation) with the extended hops.
- `recursive` → dead-end (see soundness note above).
- `unresolved` → dead-end.

Match `subject` on **type AND reporter** (decision above). `CheckReachable` sets
`Verdict` per the tier table, and sorts `Paths` deterministically (e.g. by
`Excluded` asc, then hop count, then the joined `From/Relation/To` key).

**Validation / errors** (mirror `ExplainCheck`'s policy — error only on invalid
input): unknown object type/reporter, unknown subject type/reporter, or a
relation that resolves to nothing. An **unreachable** but *valid* query is **not**
an error — it returns `Verdict: "unreachable"` with empty `Paths`.

### Step 2 — target parsing (`check.go`, next to `ParseCheckTarget`)

```go
// ParseReachTarget splits "TYPE.REPORTER#RELATION@TYPE.REPORTER" into object,
// relation, and subject. Subject is REQUIRED and must name a full facet.
func ParseReachTarget(s string) (object FacetRef, relation string, subject FacetRef, err error)
```

Reuse `ParseCheckTarget` for the `object#relation` part (split on the last `@`
first). Require the subject and require it to contain a `.` (a full facet). Keep
`ParseCheckTarget` behavior unchanged so `-check` is unaffected.

### Step 3 — formatters (new file `internal/analyze/reach_report.go`)

Mirror `check_report.go`:

- `FormatReachText(v *ReachVerdict) string` — headline verdict line
  (`✓ reachable` / `⚠ reachable only through an exclusion (unless) branch` /
  `✗ no path — schema does not support this check`), then each witness path as an
  indented hop chain (`object --relation (mult)--> ... --relation--> subject`),
  exclusion paths clearly marked. Reuse `graphdoc.Multiplicity` for cardinalities.
- `FormatReachJSON(v *ReachVerdict) (string, error)` — `json.MarshalIndent`,
  trailing newline (match `FormatCheckJSON`).

### Step 4 — tests (new file `internal/analyze/reach_test.go`)

- Unit tests over hand-built proof trees: `or` → two witnesses; `and` → both
  operands searched; `unless` → right-operand witness flagged `Excluded`; arrow
  chaining prepends hops in order; recursion terminates with no spurious witness;
  a leaf whose reporter differs from the subject reporter does **not** match
  (full-facet rule); unreachable query → empty `Paths`, `Verdict: "unreachable"`.
- Golden test over the committed schema pinning `ReachVerdict` JSON: at least one
  `reachable`, one `unreachable`, and (if the schema has an `unless`) one
  `exclusion-only` query.

**Phase-1 exit:** `go test ./internal/analyze/...` green; new goldens committed;
`-check` golden updated for the additive leaf fields.

## Phase 2 — CLI (`cmd/graph-analyze/main.go`)

- New flag `-reach TYPE.REPORTER#RELATION@TYPE.REPORTER`, honoring existing
  `-format text|json` and `-in`/`-out`. Add a `runReach` helper mirroring
  `runCheck`.
- `-check` and `-reach` are complementary; error if both are set.
- Exit code: `-reach` is an EXPLAIN-family command, **not** a CI gate — always
  exit 0 on a successful analysis (even `unreachable`), non-zero only on
  operational errors (bad input/format/write). (Contrast: the default island
  report exits non-zero on findings.)

**Phase-2 exit:** `./bin/graph-analyze -in output/graph/graph.json -reach <q>`
prints sensible text and JSON for a reachable and an unreachable query.

## Phase 3 — WASM + web wrapper (parity)

- `internal/web/reach.go`: `func CheckReachable(graph []byte, target string)
  ([]byte, error)` mirroring `internal/web/check.go` (parse target →
  `analyze.CheckReachable` → `FormatReachJSON`).
- Export `kesselCheckReach(graph, target)` in `cmd/graph-wasm/main.go` next to
  `kesselExplainCheck` — same `{ok:true, reach:"<json>"}` / `{ok:false,
  error:"..."}` envelope and the same panic-to-structured-error recovery.
- Parity golden `TestReachMatchesFilesystem` (mirror the compile/check parity
  tests): compile the committed schema, run the analysis via the CLI path and via
  `web.CheckReachable`, assert byte-identical.

**Phase-3 exit:** WASM builds (`GOOS=js GOARCH=wasm`), parity test green.

## Phase 4 — Playground UI (`internal/web`: `playground.js`, `render.js`, `playground.html`)

The UX is a **CheckRequest form**: three required selectors + a **Check** button,
a **verdict banner**, and a **path highlight** on the graph.

- **Controls** (side panel, new "Check reachability" section):
  - `#reach-object` — object facet dropdown (`TYPE.REPORTER`), populated from the
    freshly compiled graph (`graph.nodes[].reporters` → `type.reporter`).
  - `#reach-relation` — dependent dropdown: the permissions **and** relations in
    the selected object facet's scope (own + type `common` + inherited via
    `extends`). Reuse the scope rules the cost highlighter already applies
    (`relationScope`/`inheritedFacets` in `render.js`).
  - `#reach-subject` — subject facet dropdown (`TYPE.REPORTER`), **required**,
    populated from the **distinct `(Target, TargetReporter)` pairs across relation
    edges** (these are exactly the facets reachable as subjects; includes
    subject-only types like `user`).
  - `#reach-check` button labelled **"Check"** — disabled until all three are set.
- **Verdict banner** in `#reach-results`: green `✓ Path exists` / amber
  `⚠ Only via an exclusion (unless) branch` / red `✗ No path — schema does not
  support this check`. Below it, list each witness path as a hop chain (reuse the
  cost-chip styling for any per-hop annotation), exclusion paths tagged.
- **Graph highlight.** Extend the existing permission-overlay code in `render.js`
  ("Phase B", which cost-colours a clicked permission's edges) into a
  **reach-highlight mode**: fade the whole graph, then un-fade and class every
  edge on every witness path. Resolve each `WitnessHop` to its Cytoscape edge by
  matching **source facet id + relation name** within scope (same technique as
  `highlightPermission`). Add two classes:
  - `.reach-path` — an edge on at least one grant (non-excluded) witness.
  - `.reach-exclusion` — an edge appearing **only** on exclusion witnesses.
  Keep inheritance edges from the object facet visible (as `highlightPermission`
  does) so borrowed relations read as connected. Clear these classes when the
  form is reset or an empty query is submitted.
- Wire the button to `kesselCheckReach(graph, target)` (build the target string
  `object#relation@subject`), parse the JSON, render the banner + paths, and call
  the highlighter with the witness hops.

**Phase-4 exit:** in the served playground, choosing object + relation + subject
and clicking **Check** highlights the path(s) and shows the verdict; a facet pair
the schema does not connect shows the red no-path banner. Verify with the
headless-Chrome workflow (see memory `verify-web-with-headless-chrome`).

## Phase 5 — Docs (`GRAPH.md`)

Add a **"Check reachability"** section after "Check cost", matching its depth:

- Query syntax `TYPE.REPORTER#RELATION@TYPE.REPORTER`, subject required.
- The `ReachVerdict` / `WitnessPath` / `WitnessHop` contract and the verdict tiers.
- Full-facet subject matching; all-branch traversal with exclusion tagging; the
  recursion-as-dead-end soundness note.
- The graph encoding for highlighted paths (`.reach-path`, `.reach-exclusion`).
- Inherited limitations from the cost model: one-target-per-relation (a
  `typeUnion` target is not expanded), and subject-set expansion treated as a leaf
  match.
- The CLI/WASM parity guarantee, mirroring the wording used for `-check`.

## Sequencing & risk

- Phases 1 → 2 → 3 → 4 → 5 are strictly incremental; each ends shippable. Phase 1
  is the substance and is self-contained (pure functions, no UI).
- **Main correctness subtleties**, in priority order:
  1. Full-facet subject matching (type **and** reporter) — do not match on type
     alone.
  2. `unless`-right witnesses flagged `Excluded`, not dropped, and surfaced as the
     `exclusion-only` tier when they are the only witnesses.
  3. `recursive` sentinel = dead-end for witnesses (no unrolling).
  4. Witness ordering is deterministic (stable golden output).
- **Do not** re-implement name resolution or the rewrite walk — extend
  `ExplainCheck`'s leaf (Step 0) and post-walk its tree.

## Naming (avoid collision with the remote path-explorer work)

- Analysis: `CheckReachable` / `ReachVerdict` / `WitnessPath` / `WitnessHop`.
- Target parser: `ParseReachTarget`.
- CLI flag: `-reach`.
- Web wrapper: `web.CheckReachable`. WASM export: `kesselCheckReach`.
