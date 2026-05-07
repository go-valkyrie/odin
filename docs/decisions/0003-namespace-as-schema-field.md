# 0003 — Namespace as a schema field

- **Status:** Proposed (2026-05-07)
- **Deciders:** Odin maintainers
- **Consulted:** n/a

> Part of the [`v1alpha2` schema cohort](0006-v1alpha2-cohort.md).
> Cohort-wide concerns (release expectations, migration, GVK-gate
> rationale) live in 0006.

## Context

Odin already supports namespace handling at render time. The deployer
supplies a value with `odin template --namespace <ns>`, and bundle authors
expose a CUE evaluation tag — typically `_namespace: string @tag(namespace,
var=namespace)` — to receive the value and interpolate it into resource
metadata. Compat level 1 made `--namespace` no longer fail when a bundle
does not declare such a tag.

The tag-based mechanism works, but it has the same limitations the
TagVar approach has for any other identity-shaped field:

- Tags live outside the unification surface, so they are invisible to
  `odin show values`, component documentation, and any tooling that
  introspects the bundle schema.
- The value is evaluation-global. There is no clean way for a parent
  bundle to set different namespaces for different sub-bundles using the
  composition mechanism described in ADR 0001.
- Authors have to remember the somewhat idiosyncratic `@tag(name, var=name)`
  shape — discoverability through the schema is poor.

ADR 0002 introduced a pattern for promoting an identity-shaped field
(`metadata.instance`) from out-of-band to first-class: declare it on
`_#ObjectMeta`, wire bundle-level values down to components in the
internal schema, and have the loader inject the deployer-supplied value
via a presence-and-concreteness check. Namespace fits the same shape.

## Problem statement

We need a first-class schema field for namespace so that:

- Bundle authors can read it from the bundle's CUE evaluation in the same
  way they read any other piece of identity metadata.
- Tooling that walks the schema (`odin show values`, component docs,
  composition tooling) sees namespace alongside name and instance.
- Composition can scope namespace per sub-bundle without an evaluation-
  global value.
- The change rides in the `v1alpha2` cohort (see ADR 0005) for the same
  composition-gate reason as the other schema additions in the cohort.
- Existing bundles using the `@tag(namespace)` mechanism continue to work
  unchanged.

## Non-goals

- **Removing the tag-based mechanism.** The tag continues to work in this
  ADR. Its eventual retirement is acknowledged but is gated on the project
  having a deprecation policy (see open questions).
- **Cross-namespace rendering inside a single bundle.** The schema field
  carries one namespace per bundle (or per sub-bundle, in composition).
  Bundles that need to emit resources into multiple namespaces in a single
  render are out of scope here.
- **Automatic injection of namespace into resource metadata.** Mirroring
  0002, Odin does not rewrite resource fields. Authors decide how
  `metadata.namespace` flows into the resources they emit.

## Forces and constraints

- **Coexistence with the tag mechanism.** Both paths must work. Bundles
  built before this ADR continue to use the tag and are unaffected.
- **Closedness.** `_#ObjectMeta` is a closed CUE definition; bundles built
  against earlier `api/v1alpha1` versions do not have the field and must
  not break.
- **Composition compatibility.** A parent bundle may concretize a sub-
  bundle's namespace via composition (per ADR 0001). Loader injection
  must respect existing concrete values rather than overwrite them.
- **Symmetry with `instance`.** The two fields are conceptually parallel;
  authors should not have to learn two different injection models for two
  identity-shaped fields.

## Decision

### Schema change

Add `namespace` to `_#ObjectMeta` in `api/v1alpha1`, alongside the
`instance` field introduced by ADR 0002:

```cue
_#ObjectMeta: {
    name:      string
    instance:  *null | string
    namespace: *null | string
}
```

The default is `null` to distinguish "no namespace was supplied" from any
string value (including empty). Both `#Bundle` and `#ComponentBase` embed
`_#ObjectMeta`, so this single addition makes `metadata.namespace` available
at both levels.

### Internal schema wiring

Mirror the `instance` wiring: each component's `metadata.namespace` defaults
to unify with the bundle's `metadata.namespace`. A single namespace supplied
to a bundle flows to every component automatically; per-component or per-
sub-bundle overrides via composition remain possible because unification
respects already-concrete values.

### Runtime injection

The loader sets `metadata.namespace` from the deployer-supplied value (the
existing `--namespace` flag on render-path commands). The injection follows
the same presence-and-concreteness check as 0002:

1. If the field is not declared in the bundle's schema, skip — the bundle
   is built against an older `api/v1alpha1` and does not see the field.
2. If the field is declared and has not yet been concretized, inject the
   deployer value.
3. If the field has already been concretized (e.g. a parent bundle pinned
   a sub-bundle's namespace via composition), leave it alone.

### Coexistence with `@tag(namespace)`

Both paths are populated from the same `--namespace` flag. The schema field
is the new canonical surface; the tag mechanism is preserved for backward
compatibility with bundles that pre-date this ADR or have not yet been
migrated.

In bundles that use both, the values agree by construction (same source).
A bundle may use one, the other, or both; Odin does not police the choice.

### Authoring guidance

`metadata.namespace` is the first-class way to consume the value going
forward. Authoring docs should steer new bundles toward the schema field
and away from the tag.

Unlike `metadata.instance`, namespace does not feature in the canonical
`commonLabels` set (the standard `app.kubernetes.io/*` recommended labels
do not include namespace), so 0002's `commonLabels` helper is unchanged
and no new helper is introduced. Authors interpolate `metadata.namespace`
into resources the same way they interpolate `metadata.name`.

## Consequences

### Positive

- **Symmetry with `instance`.** One injection model for both
  identity-shaped fields. Authors who learn one know the other.
- **First-class visibility.** Namespace shows up in `odin show values`,
  component documentation, and any future schema-walking tooling. The
  tag mechanism's invisibility to introspection is removed for new
  bundles.
- **Composition-aware.** Parent bundles can scope a sub-bundle's namespace
  via the same unification mechanism as any other field. The
  evaluation-global limitation of the tag goes away for bundles using
  the schema field.
- **Lands cleanly in the v1alpha2 cohort.** Additive in isolation;
  ships under `v1alpha2` so composition can rely on the GVK as a
  cross-version gate (see ADR 0005). Bundles still on `v1alpha1` keep
  using the tag-based mechanism unchanged.
- **Forward-looking.** Establishes the schema-field surface that the tag
  mechanism can eventually be deprecated in favor of.

### Negative

- **Two valid paths during the transition.** Until the tag mechanism is
  deprecated, bundles can use either, both, or neither. Reviewing a
  bundle requires understanding which path it has chosen.
- **No effect on old-schema bundles.** A bundle whose schema predates the
  field is unaffected by the new mechanism — it must continue to use the
  tag. As with `--instance` in 0002, the loader needs an explicit policy
  for whether `--namespace` against an old-schema bundle that does not
  use the tag either should error, warn, or silently no-op (this is
  already the compat-level-1 question).

### Neutral

- Bundles that do not declare the tag and do not reference
  `metadata.namespace` continue to ignore the deployer-supplied value.
  This is consistent with current compat-level-1 behavior.

## Alternatives considered

### Replace the tag mechanism in this ADR

Remove `@tag(namespace)` handling immediately and require all bundles to
adopt the schema field.

Rejected because it breaks every bundle that currently uses the tag.
The schema field is additive; the tag's retirement is independent and
should follow whatever deprecation policy the project adopts.

### Have the tag set the schema field automatically

When the loader sees a `@tag(namespace)` declaration, also set
`metadata.namespace` to the same value, so authors get both paths
populated regardless of which they chose.

Tempting because it gives bundles using the tag the introspection
benefits of the schema field for free. Deferred rather than rejected:
this becomes a useful migration aid once a deprecation policy is in
place, but as a default in v1 it adds an implicit behavior that may
surprise authors who deliberately chose the tag-only path.

### Treat namespace as values, not metadata

Put `namespace` under the bundle's `values` block instead of `metadata`.

Rejected because namespace is identity (where this thing lives), not
configuration (how this thing is parameterized). It belongs alongside
`name` and `instance`, which is `_#ObjectMeta`.

## Open questions

- **Deprecation policy.** This ADR positions the schema field as the new
  canonical surface and the tag mechanism as legacy, but does not
  schedule the tag's removal. The project does not currently have a
  documented deprecation policy. A policy is needed before any
  removal can be planned, and is worth establishing as a separate
  decision (timing, compat-level interaction, communication channel).
- **Constraint on the value.** Should the API constrain `namespace` to
  DNS-1123 label form so it is always safely interpolatable into
  resource metadata, or leave it permissive and let bundle authors
  constrain per bundle? Same question is open for `instance` in 0002.
- **Behavior on old-schema bundles.** When `--namespace` is supplied
  against a bundle that does not use the tag and whose schema predates
  this field, should the runtime error, warn, or silently no-op? This
  is the same shape of question 0002 raises for `--instance` and
  arguably should be answered uniformly.

## References

- Bundle schema: `api/v1alpha1/bundle.cue`, `api/v1alpha1/component.cue`
- Internal schema wiring: `internal/schema/bundle.cue`
- Compat level 1 behavior for `--namespace`: `docs/COMPAT.md`
- Sibling ADR (multi-instance, parallel pattern): `0002-multi-instance-support.md`
- Sibling ADR (composition interactions): `0001-bundle-composition.md`
