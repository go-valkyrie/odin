# 0005 — Bundle defaults layering

- **Status:** Proposed (2026-05-07)
- **Deciders:** Odin maintainers
- **Consulted:** n/a

> Part of the [`v1alpha2` schema cohort](0006-v1alpha2-cohort.md).
> Cohort-wide concerns (release expectations, migration, GVK-gate
> rationale) live in 0006.

## Context

ADR 0001 (bundle composition) settled on a `#BundleRef` template through
which a parent bundle composes a sub-bundle. The override semantics were
described as "values-primary with a CUE unification escape hatch": the
parent overrides specific sub-bundle values either by setting them in
its own `values` block or by unifying directly onto the imported bundle
value at the import site.

Two facts about CUE make that wording insufficient for the use case
that motivates composition in the first place:

1. **CUE defaults do not compose across unification layers.** A field
   declared as `string | *"foo"` and a field declared as `string | *"bar"`
   unified together produce `string` with no default — the conflicting
   defaults cancel each other rather than one winning. This means a
   parent bundle cannot say "I want a different default than my
   sub-bundle's, but downstream deployers can still override mine."
   They can only commit to a concrete value (which downstream cannot
   override without unification conflict) or accept the sub-bundle's
   default unchanged.

2. **CUE definition closedness** (already a constraint in this codebase,
   captured elsewhere). Any schema-level mechanism for layered defaults
   must either work within closed definitions or open them up.

The motivating use case is the standard "org default override" pattern:

- Sub-bundle author publishes `webapp` with `image: string | *"nginx:latest"`
- Org A wraps it as `webapp-org-a` with their own default of
  `myorg/internal-nginx:v1.2`, expecting deployers can still override
- Deployer sets `image: "myorg/internal-nginx:v3.0"` for a specific
  cluster

CUE alone cannot express this three-layer pattern. Either Org A commits
to a concrete value (deployer override fails), or Org A accepts the
sub-bundle default (org policy doesn't apply by default).

The earlier WORK.md notes sketched an out-of-band defaults file applied
by the loader. That sketch is one of the options below; it is no longer
preserved as an in-flight design once this ADR settles.

## Problem statement

Define a defaults mechanism that supports multi-layer composition such
that:

- Sub-bundle authors set sane defaults that flow through unless
  overridden.
- Parent bundles can set their own defaults (overriding sub-bundle's)
  that downstream deployers can still override.
- Bundle authors can author defaults inline in CUE or via a separately
  shipped data file (YAML/JSON).
- The loader's role is bounded — it should not need to walk import
  graphs or know the provenance of every nested bundle.
- The mechanism integrates with the `#BundleRef` composition model
  established in ADR 0001 without requiring a new composition primitive.

## Non-goals

- **Replacing `values` as the deployer-supplied configuration surface.**
  `values` remains the top of the override stack; this ADR adds layers
  beneath it.
- **Eliminating CUE `*value` defaults entirely.** They have legitimate
  uses outside of bundle config (e.g. type definitions). The question
  is whether they remain canonical for *config* defaults.
- **Solving every authoring ergonomic.** Some patterns (partial
  override of a sub-bundle's defaults via comprehension) will be
  meaningfully more verbose than `*value` syntax. We accept this in
  exchange for compose-correctness.

## Forces and constraints

- **CUE defaults don't compose multi-layer.** This is a hard CUE
  property, not a workaround target.
- **#BundleRef provenance is not visible to the loader.** Per 0001,
  `#BundleRef` carries an imported `#Bundle` value. By the time the
  loader sees the composed bundle, it cannot natively identify which
  parts came from which sub-bundle's source location.
- **Schema additions are subject to closedness.** A new field on
  `#Bundle` is an addition to the closed definition — either additive
  within `v1alpha1` or a justification for a GVK bump.
- **Two surfaces tension.** Defaults could live as a CUE field
  (composes via unification, but inherits CUE's compose-doesn't-work
  problem if expressed via `*value`) or as out-of-band data (composes
  via merge, but loses CUE's structural composition benefits). The
  options below take different positions on this axis.

## Options considered

### Option A — Status quo, CUE `*value` defaults only

Templates and bundles use `field: type | *defaultValue` syntax. Composition
relies on CUE unification. Parents that need to override a sub-bundle's
default commit to concrete values (losing downstream overridability).

**Pros**

- No new schema, no new loader machinery.
- All defaults handling is in CUE; one mental model.

**Cons**

- The motivating use case (multi-layer override-as-default) is not
  expressible. Parent bundles must choose between accepting the
  sub-bundle's default or committing to concrete.
- The CUE compose-doesn't-work behavior is silent: an author writing
  `string | *"foo"` in a parent over a sub-bundle's `string | *"bar"`
  produces `string` with no default, which then renders as an
  unconcretized error far from the source of the problem.

### Option B — Out-of-band defaults file (WORK.md sketch)

Each bundle ships a `defaults.yaml` (or similar) alongside its source.
The loader walks the import graph, finds each imported `#Bundle`
package's defaults file, and applies them in layered order:
deepest-sub-bundle defaults → … → top-level bundle defaults → deployer
values.

**Pros**

- Defaults are plain structured data; deep-merge composition is
  uncomplicated.
- Clear separation between authored defaults and CUE-level constraints.
- No CUE schema changes (the file is purely a loader concern).

**Cons**

- **Provenance problem.** The loader has to walk the CUE import graph
  to find each sub-bundle's source location, then locate the defaults
  file alongside. Mechanically possible but brittle: post-evaluation
  CUE values do not have a clean "this struct came from package X"
  pointer; the loader has to do the lookup pre-evaluation via the
  import graph and then map paths in the composed tree to packages.
- Two surfaces for bundle-author intent: CUE source for structure and
  YAML for defaults. The loader has to coordinate them.
- Composition sophistication is limited to deep-merge — a parent that
  wants to pull part of a sub-bundle's defaults via a comprehension
  cannot, because the merge is purely loader-side.

### Option C — In-bundle `defaults` field

Add a `defaults` field to `#Bundle` (and optionally `#ComponentBase`)
holding default values as a CUE struct that mirrors the shape of
`values`. Composition of defaults happens in CUE: parent bundles unify
or comprehend over sub-bundle `defaults` to produce their own. The
loader, post-evaluation, walks the composed `defaults` tree and fills
in any non-concrete leaf in `values` with the corresponding default.

```cue
#Bundle: {
    metadata: ...
    components: ...
    values:   {...}
    defaults: {...}  // mirrors values shape
}
```

Authors write defaults inline:

```cue
defaults: components: myapp: { image: "nginx:latest", replicas: 1 }
```

…or via embed for those who prefer a YAML file:

```cue
import "encoding/yaml"

_defaultsYAML: _ @embed(file="defaults.yaml")
defaults: yaml.Unmarshal(_defaultsYAML)
```

Sub-bundle composition is an explicit author choice, ranging from full
passthrough to selective override:

```cue
// Pass sub-bundle defaults through entirely
defaults: components: base: webappbase.defaults

// Override one path, keep the rest
defaults: components: base: webappbase.defaults & {
    components: app: image: "myorg/app"
}

// Pull a subset via comprehension
defaults: components: base: components: {
    for k, v in webappbase.defaults.components if k != "secret" { "\(k)": v }
}
```

**Sub-option C1** — `defaults` on `#Bundle` only. Component templates
keep using whatever default mechanism they like; bundle author wires
sub-bundle defaults through manually.

**Sub-option C2** — `defaults` on `#Bundle` and `#ComponentBase`.
Internal-schema wiring rolls component-level defaults up into the
bundle-level `defaults` tree, mirroring how `config` flows from
`values.components[Name]`. Component templates ship canonical defaults
via this mechanism. The `*value` syntax for *config* defaults is
deprecated in favor of `defaults` blocks (see Cross-cutting sub-decisions).

**Pros (Option C generally)**

- **CUE-native composition.** Composition happens via CUE expressions
  the author writes — passthrough, unification, comprehension — without
  the compose-doesn't-work limitation, because the author controls how
  layers combine rather than relying on CUE's automatic default
  resolution.
- **No provenance problem.** The composed `defaults` tree is just CUE
  data by the time the loader sees it. The loader's job is a
  post-evaluation tree walk; it does not need to know what came from
  where.
- **Flexible authoring.** Inline CUE for short defaults, `@embed` for
  YAML/JSON files, comprehension for sophisticated cases. Authors
  pick the right tool per bundle.
- **Single source of truth per layer.** Each bundle's `defaults` block
  is the canonical place to look for "what does this bundle default?"

**Cons (Option C generally)**

- **Author burden for partial override.** "Take everything except this
  one field" requires a comprehension dance instead of `*value` syntax.
  Real ergonomic cost compared to single-layer cases, justified by the
  multi-layer correctness gain.
- **Two default systems coexist.** CUE `*value` and the new `defaults`
  block both look like "defaults" to authors. Cross-cutting sub-decision
  below addresses this.
- **Schema growth.** New field on `#Bundle` (and possibly
  `#ComponentBase`) is additive but visible.

**C1 vs C2**

- C1 keeps the schema change minimal but pushes wiring effort to bundle
  authors for every sub-bundle.
- C2 lets template authors ship canonical defaults via the new
  mechanism and gives the internal schema a place to roll them up
  automatically. Bigger schema change but better ergonomics.

### Option D — Augment composition with provenance

Extend `#BundleRef` (or revisit Option 4 from ADR 0001) so the loader
has provenance for every sub-bundle. Combine with out-of-band defaults
files as in Option B.

**Pros**

- Defaults can be plain structured data, applied per-layer by the
  loader.
- Eliminates the provenance ambiguity that made Option B brittle.

**Cons**

- Revisits ADR 0001's decision. Either adds a `#source` field to
  `#BundleRef` (redundant with the CUE import) or swaps to Option 4
  loader-orchestrated children entirely.
- Significant Go-side machinery; pulls composition logic out of CUE
  and into the loader.
- Two-surface tension remains; defaults file lives outside CUE.

### Option E — Accept the limitation

Document explicitly that CUE defaults do not compose multi-layer.
Sub-bundle authors set sane defaults; parents either accept them or
commit to concrete values. The "org default that's still overridable"
pattern is not supported.

**Pros**

- Smallest possible change; no new mechanism to design or implement.
- Honest about CUE's limitations; defers solving the problem until
  there is more pressure.

**Cons**

- The pattern that motivates composition for many real users
  (org-level customization with downstream overridability) does not
  work. Authors will hit this immediately and have no good answer.
- Pushes the problem onto every bundle author to discover and work
  around individually.

## Cross-cutting sub-decisions

### Loader algorithm

If we go Option C, the loader's role is bounded:

1. After CUE evaluation, walk the bundle's `defaults` tree.
2. For each leaf path P in `defaults`, look up `values[P]`.
3. If `values[P]` is non-concrete (`cue.Value.IsConcrete()` returns
   false), set `values[P]` to `defaults[P]` via `FillPath` or an
   overlay unification.
4. After the walk, all values that had a corresponding default are
   concrete; any remaining non-concrete values are deployer-required.

Implementation note: this walk is idempotent and uses only the bundle's
own evaluated value; it does not need access to the import graph.

### Interaction with CUE `*value` defaults

Today, component templates use `field: type | *default` for config
defaults. After Option C, two mechanisms exist for "default value of a
config field": the `*value` syntax in the template's CUE, and the
template/bundle's `defaults` block.

The cleanest answer: `defaults` blocks become canonical for *config*
defaults; `*value` remains legitimate for non-config uses (type
definitions, internal schema constraints, etc.). Existing templates
that use `*value` for config defaults can either coexist (loader
treats CUE-default-selected values as concrete, so `*value` defaults
shadow `defaults` block entries silently) or be migrated.

Under the v1alpha2 cohort decision (see below), `*value` for config
defaults is deprecated — `v1alpha2` templates use `defaults` blocks
instead. The deprecation is a clean break rather than a soft
recommendation because the GVK gate prevents v1alpha1 templates from
participating in v1alpha2 composition; the two systems do not need to
coexist within a single composed bundle.

### v1alpha2 cohort

This ADR ships under the `v1alpha2` cohort. The cohort relationship,
the GVK-gate rationale, and migration notes live in
[0006 — `v1alpha2` schema cohort](0006-v1alpha2-cohort.md); this
section covers only what is specific to defaults layering.

The `*value` deprecation for config defaults (see *Interaction with
CUE `*value` defaults* above) is the change in this ADR most directly
enabled by the cohort label. Without the GVK gate, deprecating
`*value` would be a silent recommendation in `v1alpha1` and the
loader would have to shadow `defaults` block entries with
CUE-default-selected values forever. The cohort bump turns the
deprecation into a clean break — `v1alpha2` templates use `defaults`
blocks, `v1alpha1` templates keep `*value`, and the two never coexist
inside the same composed bundle because composition is gated by GVK.

## Tentative direction

The cohort lands as `v1alpha2`. Within that:

- **Option C2** for the mechanism: `defaults` on both `#Bundle` and
  `#ComponentBase`, with internal-schema wiring rolling component
  defaults up to bundle-level. Composition happens in CUE; loader
  applies post-evaluation.
- **Deprecate `*value` for config defaults** in templates, in favor
  of the `defaults` block. Made a clean break rather than a soft
  recommendation by the `v1alpha2` GVK label.
- **Cohort-tagged with ADRs 0001-0004.** All five ADRs ship together
  under `v1alpha2`, with no expectation that any subset ships against
  `v1alpha1`.

## Open questions

- **`*value` deprecation enforcement.** The cohort GVK gates
  `v1alpha1` bundles out of `v1alpha2` composition, so `*value` for
  config defaults effectively does not coexist with the new mechanism
  in `v1alpha2`. Open: do we lint or otherwise actively flag `*value`
  on a config field in a `v1alpha2` template, or rely on authoring
  guidance alone?
- **Exact shape of the `defaults` field.** Mirror `values` exactly,
  or use a different shape that better captures defaults-specific
  concerns (e.g. nested merge rules, conditional defaults)?
- **`@embed` ergonomics.** What is the recommended pattern for
  YAML-loaded defaults? CUE's embed support has rough edges; we may
  want to document a canonical pattern, or even ship a small helper
  package in the api module.
- **Sub-bundle composition default-passthrough.** Is the explicit
  comprehension-or-unification approach the right ergonomic, or
  should `#BundleRef` auto-passthrough sub-bundle defaults into the
  parent's `defaults` tree by default? Auto-passthrough simplifies
  the common case but takes control away from the parent.

## Consequences

To be populated once the open questions above are settled. The Tentative
direction commits the cohort to `v1alpha2` and the mechanism to Option
C2; the remaining sub-decisions shape the consequences.

## Alternatives considered

Covered in the Options section above. Briefly:

- A (status quo) is rejected because it does not solve the motivating
  use case.
- B (out-of-band file) is rejected because of the loader provenance
  problem under `#BundleRef`.
- D (augment composition) is rejected because it revisits ADR 0001's
  composition decision in service of a problem that Option C solves
  more cleanly.
- E (accept the limitation) is held in reserve as the fallback if the
  ergonomic costs of Option C prove higher than expected.

## References

- ADR 0001 (composition; this ADR amends 0001's "values-primary with
  CUE unification escape hatch" wording — see Tentative direction)
- ADR 0002, 0003, 0004 (sibling schema additions; potential
  v1alpha2 cohort)
- `api/COMPAT.md` (api schema evolution rules; relevant to the
  v1alpha2 question)
- CUE documentation on defaults and unification semantics
- Earlier `WORK.md` sketch (superseded by Option B in this ADR)
