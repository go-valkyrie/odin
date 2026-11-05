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

ADRs 0001 through 0005, plus 0007, evolve the bundle schema across five
intersecting concerns:

- **Composition** (0001) — `#BundleRef` lets a bundle import another
  bundle as a component slot.
- **Identity** (0002, 0003) — `metadata.instance` and
  `metadata.namespace` as first-class schema fields, with loader-side
  injection.
- **Labeling** (0004) — `selectorLabels`/`commonLabels` helpers,
  `metadata.partOf`, `bundleMetadata` (all helpers under the component's
  `extension` struct, per 0007).
- **Defaults** (0005) — `defaults` block, deprecation of `*value` for
  config defaults.
- **Wiring** (0007) — standard `ref`, `traits`, `extension` and
  `implements` fields on `#ComponentBase`, plus standard `#Trait` and
  `#Interface` types.

Each individual change is additive in isolation. They could in
principle ship one at a time as minor versions of the api module
within `v1alpha1`. The cohort exists because composition (0001) makes
the schema *generations* matter more than individual additions:

- A `v1alpha2` `#BundleRef` expects to find the new fields
  (`metadata.instance`, `metadata.namespace`, `metadata.partOf`,
  `extension.bundleMetadata`, `defaults`) on the value its `#bundle` field is
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
| 0004 | [Standard recommended labels](0004-standard-recommended-labels.md) | Labeling helpers + `partOf` field + `extension.bundleMetadata` |
| 0005 | [Bundle defaults layering](0005-bundle-defaults-layering.md) | `defaults` block; primary GVK-gate rationale lives here |
| 0007 | [Standard component wiring: handles, traits and extensions](0007-component-handles-and-traits.md) | `ref`, `traits`, `extension`, `implements` fields; `#Trait`, `#Interface` |

Each constituent ADR is independently consequential and stands on its
own merits within the cohort. None are blocked on the others'
acceptance, but all ship as one release of the api module.

## Decision

ADRs 0001–0005 and 0007 ship together as `api/v1alpha2`. The api module retains
`api/v1alpha1` as a parallel path for bundles built before this
generation. The Go loader uses branching decoders dispatched on
`apiVersion` to handle both.

`api/v1alpha2` is authored at CUE language version v0.18 and enables
CUE's `functions` experiment (used by 0002's `resourceName` helper). Odin's
`cuelang.org/go` dependency moves to v0.18 with it; `api/v1alpha1` keeps
its current language version. Depending on an experiment is consistent
with an alpha API: if CUE changes the experiment's semantics, `v1alpha2`
follows, and the next GVK bump is where that settles.

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

## Combined schema sketch

The constituent ADRs each sketch their own part of `api/v1alpha2`. This
section puts them together, so the total size of `#ComponentBase` can be
reviewed as one decision. It was evaluated on CUE v0.18.0-rc.3 at
language version v0.18.0, together with a sample platform module and
bundle (see *Verified behavior*). Comments name the ADR each part comes
from.

```cue
@experiment(functions)

package v1alpha2

import "strings"

_#ObjectMeta: {
	name:      string
	instance:  *null | string // 0002
	namespace: *null | string // 0003
	partOf:    *null | string // 0004
}

// 0002: names produced by resourceName are DNS-1123 labels.
#Name:             string & =~"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$" & strings.MaxRunes(63)
#ResourceNameFunc: func(suffix: string = "") -> #Name

// 0007: an interface kind shared by several templates.
#Interface: {
	apiVersion: string
	kind:       string
	extension?: {...} // extension points implementers must expose
	config?: {...}    // config fields implementers must accept
}

// 0007: a trait targets a kind or interface and contributes through apply.
#Trait: {
	target: {
		apiVersion: string
		kind:       string
		metadata:   _#ObjectMeta // filled in with the target's metadata
		...
	}
	apply: {
		config?:   {...}
		defaults?: {...} // 0005 option C2
		resources?: [string]: {...}
		extension?: {...}
	}
	... // trait parameters
}

// 0002, 0004: identity helpers every component's extension carries.
// Written once as a plain struct: the base embeds it, and #Extension
// wraps it for templates that close their extension (0007).
_extensionHelpers: {
	let E = self
	bundleMetadata:    _#ObjectMeta
	componentMetadata: _#ObjectMeta
	resourceName: *(#ResourceNameFunc & func(suffix: string = "") -> #Name: strings.Join([
		if E.componentMetadata.instance != null {E.componentMetadata.instance},
		E.componentMetadata.name,
		if suffix != "" {suffix},
	], "-")) | #ResourceNameFunc
	selectorLabels: {[string]: string} | *{
		"app.kubernetes.io/name":      E.bundleMetadata.name
		"app.kubernetes.io/component": E.componentMetadata.name
		if E.componentMetadata.instance != null {"app.kubernetes.io/instance": E.componentMetadata.instance}
	}
	// Labels added by traits or bundles; merged into commonLabels.
	extraLabels: {[string]: string}
	commonLabels: {[string]: string} | *(E.selectorLabels & E.extraLabels & {
		"app.kubernetes.io/managed-by": "odin"
		if E.componentMetadata.partOf != null {"app.kubernetes.io/part-of": E.componentMetadata.partOf}
	})
}
#Extension: {_extensionHelpers}

#ComponentBase: {
	let C = self
	apiVersion: string
	kind:       string
	metadata:   _#ObjectMeta
	config: {...}              // open; templates may assign a definition
	defaults: {...}            // 0005 option C2: mirrors config
	resources: [string]: {...}
	ref?: _                    // 0007
	implements: [...#Interface] // 0007
	extension: {_extensionHelpers, ...} // 0002, 0004, 0007
	extension: componentMetadata: C.metadata

	// 0007: traits match the component's own GVK or an implemented interface.
	traits: [...#Trait & {
		target: or([
			{apiVersion: C.apiVersion, kind: C.kind, ...},
			for i in C.implements {apiVersion: i.apiVersion, kind: i.kind, ...},
		])
		target: metadata: C.metadata
	}]
	for t in traits {t.apply...}

	// 0007: conformance with implemented interfaces.
	for i in implements {
		if i.extension != _|_ {extension: i.extension}
		if i.config != _|_ {config: i.config}
	}
}

#Bundle: {
	apiVersion: "odin.go-valkyrie.com/v1alpha2"
	kind:       "Bundle"
	metadata~M: _#ObjectMeta
	components~C: [string]~(Name, _): #ComponentBase & {
		metadata: instance:  M.instance  // 0002
		metadata: namespace: M.namespace // 0003
		config: values.components[Name]
		extension: bundleMetadata: M // 0004
	}
	values: {
		components: [string]: {...}
		components: {for name, _ in C {(name): {...}}}
		...
	}
	// 0005 option C2: component defaults roll up; the loader fills
	// non-concrete values from this tree after evaluation.
	defaults: {
		components: {for name, c in C {(name): c.defaults}}
		...
	}
}
```

`#BundleRef` (0001) is not included: its shape depends on 0001's open
source-mechanism decision.

### What `#ComponentBase` contains

| Field | From | Written by | Notes |
| ----- | ---- | ---------- | ----- |
| `apiVersion`, `kind`, `metadata` | v1alpha1 | template, bundle | `metadata` gains `instance`, `namespace`, `partOf` |
| `config` | v1alpha1 | values, traits | open; a template may close it |
| `resources` | v1alpha1 | template, traits | the only field Odin renders |
| `defaults` | 0005 | template, traits | mirrors `config` |
| `ref` | 0007 | template | optional handle |
| `implements` | 0007 | template | matching and conformance |
| `extension` | 0002, 0004, 0007 | schema, template, traits, bundle | helpers plus extension points; open; a template may close it |
| `traits` | 0007 | bundle | applied by the schema |

Eight fields, up from four in `v1alpha1`. The identity helpers
(`bundleMetadata`, `componentMetadata`, `resourceName`, `selectorLabels`,
`extraLabels`, `commonLabels`) live inside `extension` rather than adding
six more.

### Pieces the constituent ADRs left open

The sketch fills in three details that the constituent ADRs leave to
the implementing change:

- **Trait matching** is `or()` over the component's own GVK and those of
  its `implements` list. Each disjunct ends with `...`, so the target's
  metadata can be filled in.
- **`componentMetadata`** is a helper field the base fills with the
  component's own metadata. It lets the helpers live in a reusable
  struct rather than referring to `C` directly.
- **Values wiring** keeps today's pattern from `internal/schema`: every
  component gets a `values.components` entry, so
  `config: values.components[Name]` always resolves.

### Verified behavior

A sample platform module (an interface, a Deployment that implements
it, a ConfigMap exposing an env trait, a Gateway with a `ref`, an
HTTPRoute, and stand-alone PDB, security-defaults and team-label traits)
and a bundle with `instance: "blue"` rendered as the ADRs describe:

- resources named `blue-api` and `blue-api-pdb` via `resourceName`;
- 0004's labels on the Deployment, plus `team: payments` from a trait
  writing `extension.extraLabels`;
- the ConfigMap's env trait injected into the Deployment's container
  through the `Workload` interface, and `securityContext` from the
  security-defaults trait;
- the PDB trait reading `replicas` and the selector labels;
- the HTTPRoute attaching to the Gateway through `ref`.

Rejected as intended: a Workload trait applied to a ConfigMap
(`target.kind: conflicting values "ConfigMap" and "Workload"`); with the
Deployment closed per 0007, `imgae` in values and `contaners` in a
trait; and a closed template claiming `Workload` without `containers`.

Found while building it, and now recorded in 0004 and 0007:

- A trait writing a label straight into `commonLabels` drops the
  identity labels when the template closes its extension (the result was
  only `managed-by` and the trait's label). `commonLabels` is a default
  disjunction, and adding a field to it is fragile under closedness. The
  sketch adds an `extraLabels` input that `commonLabels` merges in;
  traits and bundles write there, and the identity labels survive on
  open and closed templates alike.

- A trait parameter or `target` reference inside `apply` must go through
  `let T = self`. Otherwise `team: team` refers to itself.
- Function values cannot be exported to JSON, so a whole-component
  `cue export` fails on `extension.resourceName` as well as on `traits`.

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
4. Update label helpers to use `extension.selectorLabels` (in
   selectors) and `extension.commonLabels` (everywhere else) (per ADRs
   0004 and 0007).
5. Move handles, traits and extension points to the standard fields
   (per ADR 0007): `#ref` → `ref`, `#traits` → `traits`, extension-point
   definitions → `extension`, template-side trait matching →
   `implements` (with interface kinds as `#Interface`). Trait constructor
   definitions such as `#env` remain definitions.
6. Move platform modules and bundles to CUE language v0.18 and run
   `cue fix --exp=explicitopen,aliasv2`. Prefix aliases (`C=…`,
   `[Name=string]`) are not accepted at v0.18, and embedding a
   definition now closes the enclosing struct unless spread with `...`.

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
- **Mixed language versions across modules.** Can a platform module
  still at CUE language v0.17 import the v0.18 `api/v1alpha2` module, or
  must every module in a composition graph move together? Untested;
  needs a prototype before the cohort is accepted.
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
- Constituent ADRs 0001–0005 and 0007 (table above).
- Future ADR — `app.kubernetes.io/version` label semantics.
