# 0006 — `v1alpha2` schema cohort

- **Status:** Proposed (2026-05-07)
- **Deciders:** Odin maintainers
- **Consulted:** n/a

This is an umbrella ADR. It does not introduce a new decision on its
own; it documents the cohort relationship between ADRs 0001–0005 and
acts as the canonical home for cohort-wide concerns (release expectations,
migration notes, the GVK-gate rationale that previously lived inside
0005). The constituent ADRs reference this document for shared context
so they do not each have to repeat it.

## Context

ADRs 0001 through 0005 evolve the bundle schema across four
intersecting concerns:

- **Composition** (0001) — `#BundleRef` lets a bundle import another
  bundle as a component slot.
- **Identity** (0002, 0003) — `metadata.instance` and
  `metadata.namespace` as first-class schema fields, with loader-side
  injection.
- **Labeling** (0004) — `selectorLabels`/`commonLabels` helpers,
  `metadata.partOf`, `#bundleMetadata` slot.
- **Defaults** (0005) — `defaults` block, deprecation of `*value` for
  config defaults.

Each individual change is additive in isolation. They could in
principle ship one at a time as minor versions of the api module
within `v1alpha1`. The cohort exists because composition (0001) makes
the schema *generations* matter more than individual additions:

- A `v1alpha2` `#BundleRef` expects to find the new fields
  (`metadata.instance`, `metadata.namespace`, `metadata.partOf`,
  `#bundleMetadata`, `defaults`) on the value its `#bundle` field is
  bound to.
- A `v1alpha1` sub-bundle does not have those fields. CUE definition
  closedness rejects field additions, so attempting to compose a
  `v1alpha1` sub-bundle inside a `v1alpha2` parent produces a chain
  of opaque closedness errors deep in the unification tree.
- Putting the GVK string in `apiVersion` as a concrete value
  (`"odin.go-valkyrie.com/v1alpha2"`) collapses the cross-version
  case into a single, locatable unification conflict on `apiVersion`.
  That is the failure mode we want for cross-version composition:
  one clear "this bundle is the wrong version" error.

This is the **GVK-gate argument** for the cohort: composition is what
forces the homogeneity, and the GVK string is the cleanest mechanism
for surfacing version mismatches as actionable errors instead of
inscrutable closedness chains.

## Constituents

| ADR  | Title                                                       | Cohort role |
| ---- | ----------------------------------------------------------- | ----------- |
| 0001 | [Bundle composition](0001-bundle-composition.md)            | Trigger — the change that forces the cohort to exist |
| 0002 | [Multi-instance support](0002-multi-instance-support.md)    | Identity field on `_#ObjectMeta` |
| 0003 | [Namespace as a schema field](0003-namespace-as-schema-field.md) | Identity field on `_#ObjectMeta` |
| 0004 | [Standard recommended labels](0004-standard-recommended-labels.md) | Labeling helpers + `partOf` field + `#bundleMetadata` slot |
| 0005 | [Bundle defaults layering](0005-bundle-defaults-layering.md) | `defaults` block; primary GVK-gate rationale lives here |

Each constituent ADR is independently consequential and stands on its
own merits within the cohort. None are blocked on the others'
acceptance, but all ship as one release of the api module.

## Decision

ADRs 0001–0005 ship together as `api/v1alpha2`. The api module retains
`api/v1alpha1` as a parallel path for bundles built before this
generation. The Go loader uses branching decoders dispatched on
`apiVersion` to handle both.

`v1alpha1` bundles continue to:
- Load and render at the top level.
- Use the existing `@tag(namespace)` mechanism for namespace handling.
- Rely on CUE `*value` for config defaults if the template uses it.

`v1alpha1` bundles **cannot**:
- Be composed as sub-bundles via `#BundleRef` from a `v1alpha2`
  parent. The GVK gate rejects this at unification time.

`v1alpha2` is the target for any new bundle that wants composition,
multi-instance, namespace as a schema field, the labeling helpers, or
the defaults mechanism — i.e. effectively all new bundles going forward.

## Migration

Bundle authors migrating from `v1alpha1` to `v1alpha2`:

1. Bump the api module dependency in `cue.mod/module.cue`.
2. Replace `@tag(namespace, var=namespace)` declarations with direct
   references to `metadata.namespace` (per ADR 0003).
3. Move config defaults from `*value` markers to `defaults` blocks
   (per ADR 0005). Templates that ship as part of the bundle and
   define their own config schema do this in the template; bundles
   that supply default values for a sub-bundle's components do this
   in the bundle's `defaults` block.
4. Update label helpers to use `selectorLabels` (in selectors) and
   `commonLabels` (everywhere else) (per ADR 0004).

Bundles that do not adopt `v1alpha2` continue to render under
`v1alpha1` indefinitely, modulo the established `compat` levels in
`docs/COMPAT.md`. There is no forced migration timeline at this
stage — `v1alpha1` removal would be its own decision.

## Out of scope for v1alpha2

The following were considered as candidates and explicitly deferred:

- **`app.kubernetes.io/version` label.** Application-vs-bundle version
  semantics need their own decision; deferred to a future ADR.
- **Deprecation policy for the `@tag(namespace)` mechanism.** ADR 0003
  treats the schema field as the new canonical surface but does not
  schedule the tag's removal. Removal requires a project-wide
  deprecation policy, which is its own concern.
- **Auto-injection of `metadata.partOf`** from `#BundleRef`. Logical
  extension but requires composition-side wiring patterns that have
  not been exercised yet (per ADR 0004's open questions).

## Cohort-level open questions

- **When does the cohort move from Proposed to Accepted?** The
  individual ADRs each have their own open questions; this umbrella
  moves to Accepted once they all do (or at least once their open
  questions are reduced to non-blocking sub-decisions).
- **`v1alpha1` deprecation horizon.** The current decision is "support
  indefinitely." A real horizon — measured in releases or calendar
  time — should be considered once `v1alpha2` adoption is meaningful.
- **Future cohorts.** Whether subsequent generations (`v1alpha3`,
  eventually `v1beta1` and `v1`) follow the same cohort-bumping
  pattern, or whether smaller individual GVK bumps make sense once
  the api stabilizes.

## References

- `api/COMPAT.md` — schema-evolution rules; the cohort-bump trigger
  is documented there.
- Constituent ADRs 0001–0005 (table above).
- Future ADR — `app.kubernetes.io/version` label semantics.
