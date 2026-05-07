# API schema compatibility

The `go-valkyrie.com/odin/api` module is versioned via SemVer at the
major-version level. Pre-v1 (`@v0` today), the module carries no stability
guarantee — breaking changes within `@v0` are explicitly permitted. Bumping
to `@v1` is what will lock a stable contract.

Within a major version, schema evolution takes two forms:

- **Additive within a GVK** (e.g. adding a field to `v1alpha1`). Preferred
  when feasible. Consumers pick up the new shape by bumping their api dep
  within the same `@v0.x.y` line, and minimum-version selection across a
  composition graph resolves cleanly to a single api version every bundle
  can use.
- **GVK bump** (e.g. `v1alpha1` → `v1alpha2`). The right answer when:
  - A change is not additive — removing a field, renaming, changing the
    semantics of an existing field; or
  - A coherent cohort of changes lands together where the cohort's
    correctness depends on consumers seeing all of them at once. The
    canonical case is composition: when a parent bundle pulls in a
    sub-bundle, both must see the same schema or unification produces
    deep, hard-to-diagnose closedness errors. The GVK string in
    `apiVersion` becomes a CUE-level gate that surfaces version
    mismatches as a single clean unification conflict instead.

  Pre-v1, GVK bumps are a normal evolution path, not a major event.

A bundle pins a specific api module version in its `cue.mod/module.cue`.
CUE's minimum-version selection ensures every bundle in a composition
graph resolves to one api version, and within that version all schemas are
mutually consistent. Cross-version concerns — supporting v1alpha1 bundles
and v1alpha2 bundles in the same Odin install — live on the Go side as
branching decoders that dispatch on GVK.
