# 0002 — Multi-instance support

- **Status:** Proposed (2026-05-07)
- **Deciders:** Odin maintainers
- **Consulted:** n/a

> Part of the [`v1alpha2` schema cohort](0006-v1alpha2-cohort.md).
> Cohort-wide concerns (release expectations, migration, GVK-gate
> rationale) live in 0006.

## Context

An Odin bundle today produces a fixed set of Kubernetes resources whose names
are determined by the bundle author. There is no notion of *which deployment*
of a bundle a given render represents. As a result, deploying the same bundle
twice into the same namespace produces resource-name collisions: the second
render is indistinguishable from the first at the manifest level.

This is the same problem Helm solves with release names. In Helm, a chart's
templates can refer to `{{ .Release.Name }}`, and the deployer chooses a
release name per `helm install` so that resources from co-resident releases
of the same chart do not collide.

Odin already has a clean place for "what is this thing" — `metadata.name` on
both `#Bundle` (the bundle's intrinsic name, equivalent to a Helm chart name)
and `#ComponentBase` (the component's role within a bundle: `webapp`,
`database`, etc.). What it lacks is "which deployment of this thing" — the
release-name analogue.

## Problem statement

We need a per-deployment identifier such that:

- The deployer can supply it at render time (e.g. `odin template --instance
  primary ./bundle`).
- Bundle authors can read it from the bundle's CUE evaluation and template it
  into resource names and labels.
- It composes cleanly with the bundle composition mechanism described in
  ADR 0001 — a parent bundle pinning a sub-bundle's instance must take
  precedence over Odin's runtime injection.
- It rides in the `v1alpha2` cohort (see ADR 0005) so that composition
  with sub-bundles can rely on every participant seeing the same schema.

## Non-goals

- **Automatic resource-name rewriting.** Odin will not implicitly prefix or
  suffix resource names with the instance identifier. Helm does this and
  the implicit rewriting (combined with the need to track label selectors
  and cross-resource references) is a recurring source of surprises. Bundle
  authors decide how the instance flows into resource names.
- **Retrofitting bundles built against older `api/v1alpha1` versions.**
  Bundles whose schema predates this field remain single-instance. They
  continue to load and render unchanged.

## Forces and constraints

- **CUE definition closedness.** `#Bundle` and `#ComponentBase` are CUE
  definitions and reject unification with fields not declared in the schema.
  Any approach that injects an instance value through the bundle's
  unification surface must ensure the field is already declared at the
  schema version the bundle was authored against.
- **Cohort with v1alpha2.** This change is additive in isolation but
  ships as part of the `v1alpha2` cohort (composition + identity +
  labels + defaults — see ADR 0005). The cohort's GVK gate is what
  makes composition surface cross-version mismatches as a single clean
  unification error instead of a chain of closedness errors deep in
  the tree.
- **Bundle composition compatibility.** A parent bundle may concretize a
  sub-bundle's instance via composition. Runtime injection must respect
  pre-existing concrete values rather than overwrite them.
- **Discoverability.** Bundle authors should be able to tell from the
  schema whether a bundle supports multi-instance deployment.

## Decision

### Schema change

Add `instance` to `_#ObjectMeta` in `api/v1alpha1`:

```cue
_#ObjectMeta: {
    name:     string
    instance: *null | string
}
```

Both `#Bundle` and `#ComponentBase` already embed `_#ObjectMeta`, so this
single addition makes `metadata.instance` available at both levels without
duplicate declarations. The default value is `null`, which distinguishes
"no instance was supplied" from any string value (including the empty
string).

### Internal schema wiring

`internal/schema/bundle.cue` defaults each component's `metadata.instance` to
unify with the bundle's `metadata.instance`, mirroring how `config:
values.components[Name]` already flows bundle-level configuration down into
components. This means a single instance value supplied to the bundle flows
to every component automatically; per-component overrides via composition
remain possible because unification respects already-concrete values.

### Runtime injection

Odin's loader is responsible for setting `metadata.instance` from the
deployer-supplied value (a `--instance` CLI flag on render-path commands).
Injection follows a presence-and-concreteness check:

1. If the field is not declared in the bundle's schema, skip — the bundle is
   built against an older `api/v1alpha1` and is single-instance by design.
2. If the field is declared and has not yet been concretized (still defaults
   to `null` or has not been set by composition), inject the deployer value.
3. If the field has already been concretized (e.g. a parent bundle pinned a
   sub-bundle's instance via composition), leave it alone.

### Authoring helpers

`#ComponentBase` exposes two derived fields whose values default to the
canonical multi-instance conventions. Bundle and template authors can use
them directly or override them by unification when a different convention
is required.

```cue
#ComponentBase: {
    ...
    metadata: _#ObjectMeta
    resourceName: string | *[
        if metadata.instance != null { "\(metadata.instance)-\(metadata.name)" },
        metadata.name,
    ][0]
    commonLabels: {[string]: string} | *{
        "app.kubernetes.io/name": metadata.name
        if metadata.instance != null {
            "app.kubernetes.io/instance": metadata.instance
        }
    }
}
```

`resourceName` produces the conventional resource name — instance-prefixed
when an instance is set, the bare component name otherwise. `commonLabels`
produces the standard `app.kubernetes.io/{name,instance}` pair, omitting
`instance` when the bundle is single-instance. Both default branches handle
the null case so authors do not reproduce the conditional in every
component.

The fields are exposed as defaults (`*expr | T`), not enforced computations.
Authors can override at three levels by unification:

- **Per-component**: directly set `resourceName` or `commonLabels` on a
  component instance.
- **Per-bundle**: define a local `#ComponentBase` wrapper that pins a
  different convention, e.g.
  ```cue
  #ComponentBase: odin.#ComponentBase & {
      resourceName: "\(metadata.name)-\(metadata.instance)"
  }
  ```
- **Per-component-template**: a component template can pin its own
  convention when it has structural constraints (e.g. CronJob name-length
  limits).

Public field names (rather than hidden `_resourceName`) are required because
hidden identifiers in CUE are package-scoped and not visible to consumer
modules.

### Templating responsibility

Resource-name and label templating remain the bundle author's responsibility
— Odin does not rewrite either. The helpers above mean the canonical
convention is one identifier reference rather than a hand-written
conditional. Authoring guidance will document:

- The expected use of `resourceName` for resource `metadata.name` fields and
  any name-derived cross-references (selectors, owner refs, etc.).
- The expected use of `commonLabels` for resource labels and selector
  matchers.
- The override patterns above for bundles or templates that need a
  different convention.

## Consequences

### Positive

- **Lands cleanly in the v1alpha2 cohort.** The schema addition is
  additive in isolation, but ships under `v1alpha2` so composition can
  rely on the GVK as a cross-version gate. Bundles still on `v1alpha1`
  continue to load and render — they simply do not see the new field
  and remain single-instance.
- **Single point of declaration.** Adding `instance` to `_#ObjectMeta`
  covers both `#Bundle` and `#ComponentBase` with one schema edit.
- **Composition-aware by construction.** Odin's "inject only if not yet
  concrete" rule means parent-set values defeat runtime injection without
  any special casing in the loader.
- **Self-documenting.** A bundle's schema reveals whether it supports
  multi-instance deployment by virtue of declaring (or inheriting) the
  `instance` field. `odin show values` and component documentation surface
  it without bespoke handling.
- **Convention as suggestion, not mandate.** `resourceName` and
  `commonLabels` give authors convention-conforming defaults that work out
  of the box and a clean override hook (per-component, per-bundle wrapper,
  or per-template) when conventions need to deviate. Most bundles converge
  on the same convention without any author effort; the rare bundle that
  needs to deviate can do so with a one-line unification.
- **Generalizable pattern.** The presence-and-concreteness check is a
  template for any future runtime-supplied identity-adjacent context
  (cluster name, environment, etc.) — additions can ride future GVK
  cohorts in the same way this one rides `v1alpha2`.

### Negative

- **Silent override of the default helpers.** Because `resourceName` and
  `commonLabels` are exposed as defaults (`*expr | T`), an override that
  diverges from the canonical convention unifies cleanly rather than
  producing a CUE conflict. An author who intends to use the convention
  but accidentally writes a different value will not see an error. This
  is the cost of providing an escape hatch and is judged acceptable; the
  alternative — a hard-coded computation that conflicts with any divergent
  unification — closes the door on legitimate per-bundle conventions.
- **Templating outside the helpers requires manual null-guarding.** Authors
  who interpolate `metadata.instance` directly without using `resourceName`
  or `commonLabels` must guard the null case themselves; failing to do so
  produces a CUE evaluation error in single-instance mode. Authoring
  guidance should steer authors toward the helpers as the default path.
- **Silent inactivity on old schemas.** A `--instance` flag passed against
  a bundle whose schema predates the field has no effect. The loader needs
  an explicit policy (error by default, with an opt-in lenient mode) to
  avoid a confusing failure mode in which the deployer believes they have
  set an instance but the rendered output is unchanged.

### Neutral

- Bundles that do not need multi-instance support can continue to ignore
  `metadata.instance` entirely. The default of `null` means no behavior
  change for those bundles.

## Alternatives considered

### TagVar (CUE `@tag()`)

Use a CUE evaluation tag (`_instance: string @tag(instance)`) instead of a
schema field. Tags are not part of the unification surface and so sidestep
the closedness problem.

Rejected because:

- Tags are evaluation-global. A parent bundle composing two sub-bundles
  cannot give them different instance values cleanly — there is one
  `instance` tag for the whole evaluation.
- Tags are untyped (always strings), invisible to `odin show values` and
  component documentation, and offer no per-bundle constraint surface.
- A future migration from tag to first-class field would itself require
  schema work, so the approach defers rather than avoids the cost.

### GVK bump (`v1alpha2`) on this change alone

Bump the API version solely to ship `metadata.instance`, with no
sibling cohort.

Rejected because in isolation this addition is purely additive and does
not justify the cost of parallel decoders. The bump that does happen —
to `v1alpha2` — is driven by ADR 0005's cohort reasoning, not by this
change on its own.

### Field on `#Bundle` and `#ComponentBase` directly

Functionally equivalent to placing it on `_#ObjectMeta`, but duplicates the
declaration in two places and treats the field as something other than
identity metadata, which is exactly what it is.

### Automatic resource-name rewriting

Have Odin automatically prefix or suffix resource names with the instance
identifier (Helm-style implicit templating).

Rejected because Odin would need to understand label selectors, cross-
resource references, and other manifest semantics to rewrite names without
breaking inter-resource references. This is the part of Helm that has
historically been a source of bugs and surprises. Leaving templating
explicit keeps Odin's contract narrow.

## Open questions

- **Surfacing for old-schema bundles.** When `--instance` is supplied
  against a bundle whose schema predates the field, should the runtime
  error, warn, or silently ignore? The current lean is error by default
  with an opt-in lenient flag, but the right knob shape is undecided.
- **Constraint on the value.** Should the API constrain `instance` (e.g.
  to DNS-1123 label form) so that it is always safely interpolatable into
  resource names, or leave it permissive and let bundle authors constrain
  per-bundle? Permissive is simpler today but constraint at the API level
  would prevent a class of bundle-author mistakes.
- **Authoring guidance home.** Where does the templating convention live —
  CONTRIBUTING, a separate authoring guide under `docs/`, or a cookbook?

## References

- Bundle schema: `api/v1alpha1/bundle.cue`, `api/v1alpha1/component.cue`
- Internal schema wiring: `internal/schema/bundle.cue`
- Sibling ADR (composition interactions): `0001-bundle-composition.md`
