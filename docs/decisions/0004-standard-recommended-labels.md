# 0004 — Standard recommended labels

- **Status:** Proposed (2026-05-07)
- **Deciders:** Odin maintainers
- **Consulted:** n/a

> Part of the [`v1alpha2` schema cohort](0006-v1alpha2-cohort.md).
> Cohort-wide concerns (release expectations, migration, GVK-gate
> rationale) live in 0006.

## Context

ADR 0002 introduced a `commonLabels` helper on `#ComponentBase` covering
`app.kubernetes.io/name` and `app.kubernetes.io/instance`. That covered the
minimum but made one off-convention choice: `app.kubernetes.io/name` was
sourced from the *component's* `metadata.name` rather than the *bundle's*.
The Kubernetes and Helm convention is the opposite — `name` carries the
bundle/chart-level identity, and `component` carries the role within it.

Two additional things are worth correcting now while 0002 is still in
Proposed status and no bundles depend on the current shape:

- The recommended-labels set is broader than just `name` and `instance`.
  `component`, `part-of`, and `managed-by` round out the conventional
  five (six if `version` is included; see below). Adopting the full set
  with correct semantics gives bundles standardized identity for
  selectors, observability, and GitOps tooling out of the box.
- Some labels are stable across deployments (safe in `spec.selector`),
  others are informational and change between deployments. Helm
  exposes this as two helpers — `chart.selectorLabels` (subset) and
  `chart.labels` (full set). The distinction matters: putting an
  unstable label like `version` in a Deployment selector breaks
  redeployments. Bundles benefit from having the safe subset named
  separately.

`app.kubernetes.io/version` is deliberately deferred to a follow-up ADR.
The label semantically refers to the *application's* version, not the
bundle's, and they do not always align (a `:latest` bundle deploying PR
previews of an application is the canonical counter-example). Sourcing
that field correctly deserves its own decision rather than being bundled
in here.

## Problem statement

Provide standard Kubernetes recommended labels as helpers on
`#ComponentBase` such that:

- Sources are semantically correct (bundle identity in `name`, component
  role in `component`).
- The selector-safe subset is named separately from the full set so
  authors do not accidentally put unstable labels in selectors.
- The labels compose cleanly with bundle composition (ADR 0001) — each
  `#Bundle` forms its own labeling boundary, so sub-bundle resources
  carry the sub-bundle's name in `name`, not the parent's.
- The change rides in the `v1alpha2` cohort (see ADR 0005) for the
  composition-gate reasons documented there.

## Non-goals

- **`app.kubernetes.io/version`.** Deferred to a follow-up ADR. Setting
  it correctly requires a separate decision about app-vs-bundle version
  semantics.
- **Annotations.** Some recommended metadata (e.g. `helm.sh/chart`) lives
  on annotations rather than labels. Annotation helpers are out of scope
  for this ADR.
- **Enforcing labels on emitted resources.** Mirroring 0002, the helpers
  are defaults that authors apply explicitly. Odin does not rewrite
  resource metadata.

## Forces and constraints

- **Selector stability.** Anything in `selectorLabels` must be stable
  across deployments of the same logical thing. Adding informational
  labels to a Deployment selector breaks redeployment.
- **0002 revision.** 0002's helper currently uses the component's name
  for `app.kubernetes.io/name`. That is wrong for multi-service bundles
  (each component would carry a different `name` despite belonging to
  the same logical app). Since 0002 is Proposed, revising the helper
  here is fair.
- **Component access to bundle metadata.** For a helper on
  `#ComponentBase` to source `name` from the bundle, the component must
  see its bundle's `metadata` during CUE evaluation. Currently it does
  not — components are nested in the bundle but have no back-reference
  to it. Wiring this is part of the change.
- **Sub-bundle labeling boundary.** Each `#Bundle` is its own labeling
  unit. Sub-bundle components see the sub-bundle's metadata, not the
  parent's. This falls out for free if the wiring is per-`#Bundle`.
- **CUE definition closedness.** Schema changes must work with closed
  definitions: new fields on `_#ObjectMeta` are additions to the
  definition; auxiliary surfaces on `#ComponentBase` should use
  definition slots (`#foo: ...`) rather than regular fields, since
  closedness rejects ad-hoc field additions but not definitions.

## Decision

### Schema changes

Add `partOf` to `_#ObjectMeta`:

```cue
_#ObjectMeta: {
    name:      string
    instance:  *null | string
    namespace: *null | string
    partOf:    *null | string
}
```

Default null means the corresponding label is omitted unless the bundle
author or composition mechanism sets a value.

Add a definition slot `#bundleMetadata` on `#ComponentBase` so each
component can see its bundle's metadata:

```cue
#ComponentBase: {
    _#TypeMeta
    metadata: _#ObjectMeta
    config:    {...}
    resources: [string]: {...}
    #bundleMetadata: _#ObjectMeta
    ...
}
```

`#bundleMetadata` is a definition slot, not a regular field, so it
bypasses closedness when being filled in by `#Bundle`'s wiring. Definitions
also are not module-namespaced (unlike hidden fields), so external bundle
modules can populate it.

### Internal schema wiring

`#Bundle` populates each component's `#bundleMetadata` with its own
metadata:

```cue
#Bundle: {
    M=metadata: _#ObjectMeta
    C=components: [Name=string]: #ComponentBase & {
        config: values.components[Name]
        #bundleMetadata: M
    }
    ...
}
```

This wiring lives on the api `#Bundle` definition (per the rationale in
ADR 0001 — bundle wiring belongs in the api so it applies to nested
bundle values as well).

### Helpers

Replace 0002's `commonLabels` with two helpers on `#ComponentBase`.

```cue
#ComponentBase: {
    ...
    selectorLabels: {[string]: string} | *{
        "app.kubernetes.io/name":      #bundleMetadata.name
        "app.kubernetes.io/component": metadata.name
        if metadata.instance != null {
            "app.kubernetes.io/instance": metadata.instance
        }
    }
    commonLabels: {[string]: string} | *(selectorLabels & {
        "app.kubernetes.io/managed-by": "odin"
        if metadata.partOf != null {
            "app.kubernetes.io/part-of": metadata.partOf
        }
    })
}
```

`selectorLabels` is the selector-stable subset: identity that does not
change between deployments of the same logical thing. `commonLabels` is
the full informational set, building on `selectorLabels` and adding
`managed-by` (always `"odin"`) and `part-of` (when set).

`version` is intentionally absent from both helpers and will be
addressed in a follow-up ADR.

### Templating convention

Authoring guidance:

- **Selectors** (`spec.selector.matchLabels`, the selector subset of
  `spec.template.metadata.labels`) use `selectorLabels`.
- **Resource labels** (everywhere else `metadata.labels` appears) use
  `commonLabels`.

Both helpers are exposed as defaults (`*expr | T`), so authors can
override or extend by unification when convention does not fit.

### Composition behavior

Each `#Bundle` is its own labeling boundary because the wiring is per-
`#Bundle`. Concretely:

- A top-level bundle's components see the top-level bundle's metadata
  via `#bundleMetadata`. Their `app.kubernetes.io/name` is the top-level
  bundle's name.
- A sub-bundle pulled in via `#BundleRef` (ADR 0001) is itself a
  `#Bundle` value, so its components see the sub-bundle's metadata via
  the sub-bundle's own wiring. Their `app.kubernetes.io/name` is the
  sub-bundle's name, not the parent's.

This matches the Helm subchart convention: each chart, including
subcharts, is its own labeling unit.

`part-of` is the explicit pointer between layers when desired. For now,
authors set `metadata.partOf` themselves where it makes sense. Auto-
injection of the parent bundle's name into a sub-bundle's `partOf`
through `#BundleRef` is a logical extension but requires additional
wiring; see Open questions.

### Revision of 0002's helper

0002's `commonLabels` is replaced by the `selectorLabels` and
`commonLabels` defined here. The behavioral changes are:

- `app.kubernetes.io/name` now sources from the bundle's `metadata.name`,
  not the component's.
- `app.kubernetes.io/component` (new) carries the component's
  `metadata.name`.
- `app.kubernetes.io/managed-by` and `app.kubernetes.io/part-of` are new
  on `commonLabels`.
- `selectorLabels` is new; authors using the helper in selectors should
  switch to it.

0002 is in Proposed status and no bundles depend on the current shape;
the revisions can land together.

## Consequences

### Positive

- **Conventional semantics.** `name` carries bundle identity and
  `component` carries role, matching Helm and Kubernetes practice.
  Multi-service bundles get correct labels by default.
- **Selector safety.** The selector-stable subset is named separately,
  reducing the chance of redeployment-breaking label drift through
  selectors.
- **Composition-aware labeling.** Each bundle is its own labeling
  boundary, sub-bundles included. No special-casing needed.
- **Lands cleanly in the v1alpha2 cohort.** Additive in isolation;
  ships under `v1alpha2` so composition can rely on the GVK as a
  cross-version gate (see ADR 0005).
- **Build-up pattern.** `commonLabels` is `selectorLabels` plus extras,
  so authors who customize one composes naturally with the other.
- **Generalizable.** The `#bundleMetadata` slot is reusable for any
  future helper that needs bundle-level identity from a component.

### Negative

- **Two helper concepts.** Authors must learn the
  `selectorLabels`/`commonLabels` split. Documentation needs to make
  the convention explicit.
- **`#bundleMetadata` is a new public surface.** It is a definition
  slot consumers can reference, which expands the api's exposed
  vocabulary slightly.
- **Single-component bundles get redundant labels.** A bundle with one
  component carries both `name` and `component` even though they are
  effectively the same identity layer. Trivial overhead but visually
  noisy. See Open questions.
- **Manual `part-of` for now.** Sub-bundle composition does not
  auto-populate `part-of` in v1; authors who want it set per
  sub-bundle do so manually.

### Neutral

- Bundles that ignore the helpers entirely are unaffected. `commonLabels`
  is a default, not a constraint.

## Alternatives considered

### Single helper, no selector subset

Expose only `commonLabels` and let authors filter to selector-safe
labels themselves.

Rejected because it makes the unstable-label-in-selector footgun easy
to hit and harder to detect. Helm's two-helper convention exists for
exactly this reason.

### Keep 0002's name semantics

Source `app.kubernetes.io/name` from the component's `metadata.name`
and skip `app.kubernetes.io/component` entirely.

Rejected because it produces wrong labels for multi-service bundles —
each component carries a different `name` despite belonging to the same
logical app. Also makes "find me everything for app X" harder than it
should be.

### Auto-inject `part-of` from `#BundleRef`

Have `#BundleRef` wire its parent bundle's `metadata.name` into each
sub-bundle component's `metadata.partOf`.

Tempting, but requires nested wiring through composition that we have
not exercised yet. Defer to future ADR once the patterns are clearer.
Authors who want it can set `metadata.partOf` explicitly today.

### Annotations instead of labels for `managed-by`/`part-of`

Some tools use annotations for chart-level metadata
(`helm.sh/chart`).

Rejected for these specific labels. `managed-by` and `part-of` are
explicitly part of the Kubernetes recommended labels set and are
queryable via selectors when desired. Annotations are a separate
surface.

### Always emit `component`, even when redundant

Always include `app.kubernetes.io/component` on `selectorLabels`. The
proposed helper does this.

Considered: omit `component` when a bundle has only one component (the
component name equals the bundle name and the label is redundant).
Rejected because it makes the helper output non-uniform and complicates
selectors that match `component` for a known role. Visual redundancy
is acceptable.

## Open questions

- **Auto-injection of `part-of`.** Should `#BundleRef` populate the
  sub-bundle's `metadata.partOf` from its enclosing bundle's
  `metadata.name`? The wiring would mirror how `config: values.components`
  threads bundle state through, but applied across the composition
  boundary. Worth considering once we have more concrete experience
  with composition. Decision deferred.
- **Naming of the selector-stable helper.** `selectorLabels` matches
  Helm. `coreLabels` and `identityLabels` are alternatives. `selectorLabels`
  is more discoverable to readers familiar with Helm's vocabulary;
  this ADR defaults to it.
- **`#bundleMetadata` vs an alternative wiring mechanism.** A definition
  slot is the closedness-friendly option, but other approaches (an
  exported field, a hidden field that the api reflects through some
  other convention) might be cleaner. The current proposal works; if
  the wiring story for the broader composition machinery settles on a
  different pattern, this should align with it.
- **Single-component bundle redundancy.** A bundle with one component
  emits both `name` and `component` carrying identical-or-near-identical
  identity. Acceptable today but worth revisiting if the noise becomes
  meaningful in practice.

## References

- Bundle and component schemas: `api/v1alpha1/bundle.cue`,
  `api/v1alpha1/component.cue`
- Sibling ADR (multi-instance, introduced original `commonLabels`):
  `0002-multi-instance-support.md`
- Sibling ADR (composition interactions):
  `0001-bundle-composition.md`
- Sibling ADR (namespace as schema field, same v1alpha2 cohort):
  `0003-namespace-as-schema-field.md`
- Sibling ADR (defaults layering, motivates the `v1alpha2` bump):
  `0005-bundle-defaults-layering.md`
- Kubernetes recommended labels:
  https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
- Helm chart labels and selector labels conventions:
  https://helm.sh/docs/chart_best_practices/labels/
- Future ADR: `app.kubernetes.io/version` label
