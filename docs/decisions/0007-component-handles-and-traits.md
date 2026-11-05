# 0007 — Standard component wiring: handles, traits and extensions

- **Status:** Proposed (2026-11-05)
- **Deciders:** Odin maintainers
- **Consulted:** n/a

> Part of the [`v1alpha2` schema cohort](0006-v1alpha2-cohort.md).
> Cohort-wide concerns (release expectations, migration, GVK-gate
> rationale) live in 0006.

## Context

Platform modules have built a component-wiring vocabulary on top of
`v1alpha1` without any support from the Odin schema. A representative
example (lightly abridged from an internal platform module):

```cue
// core module: the shared trait shape
Trait: {
    #target: { apiVersion: string, kind: string, metadata: name: string }
    #extension: { odin.#ComponentBase }
    ...
}

// workload module: an interface kind shared by CronJob, WebApp, Deployment, ...
_#Trait: core.Trait & {
    #target: { apiVersion: "platform.vituity.com/v1alpha1", kind: "Workload" }
}

_#Workload: C=odin.#ComponentBase & {
    _allowedTraits: _#Trait | core.Trait & {#target: {apiVersion: C.apiVersion, kind: C.kind}}
    #containers: [string]: _#ContainerConfig     // extension point
    #volumes: [Name=string]: schema.#VolumeSpec  // extension point
    #traits: [...{_allowedTraits | error("invalid trait")}]
    for trait in #traits { trait.#extension }
    ...
}
```

Bundles then wire components together:

```cue
components: {
    mygateway: gateway.#Gateway & { ... }
    myroute: gateway.#HTTPRoute & { config: gateway: mygateway.#ref }
    myapp: workload.#Deployment & {
        #traits: [configmap.#env, configmap.#mount & {path: "/config"}, mysecuredefaults]
    }
}
```

Four patterns appear:

1. **Handles.** `#ref`, a value other components use to point at this
   one.
2. **Traits.** `#traits`, a list of cross-cutting extensions applied to
   a component, each carrying a target and an extension.
3. **Extension points.** `#containers`, `#volumes`: the parts of a
   template that traits write into.
4. **Interface kinds.** `Workload` is not a real component kind. It's a
   target shared by several templates, so one trait can apply to all of
   them.

Every one of these uses definition fields for one reason only:
`#ComponentBase` is closed, and CUE's closedness check applies to regular
fields but not to definitions. That workaround has real costs:

- **Silent typos.** A trait writing `#contaners` adds a new definition
  instead of failing.
- **Invisible to tooling.** `_allowedTraits` is hidden and
  package-scoped. Neither Odin nor a GUI can see which traits a template
  accepts without trying to unify each candidate.
- **No cross-module contract.** Traits are used across modules, but the
  shared shape only exists by convention in one platform module.

Odin already has partial awareness of this surface: `@odin(ref)`,
`@odin(ext)` and `@odin(other)` annotations mark public declarations for
`odin docs`.

`v1alpha2` is already a GVK bump, so standardizing now costs one
migration note. Doing it later would cost another bump.

## Problem statement

Standardize handles, traits, extension points and trait matching on
`#ComponentBase`, so that:

- typos and mismatches are schema errors;
- which traits a component accepts can be determined from data, without
  evaluation;
- traits work across modules;
- traits can read their target's metadata and config, and contribute to
  its `config`, `resources` and extension points;
- interface kinds can declare what templates implementing them must
  expose;
- platform modules keep control of what extension points and interface
  kinds exist.

## Non-goals

- **Go-side trait semantics.** Trait application is expressed in the
  api module's CUE. The loader does not apply, order or merge traits.
- **A standard catalog** of traits, extension points or interface kinds.
  Those remain platform-module territory.
- **Replacing trait constructors.** Definitions such as
  `configmap.#env` stay definitions published with `@odin(ext)`. Only
  the fields they flow through are standardized.

## Forces and constraints

- **Keep `#ComponentBase` small.** Every standard field is a pattern
  Odin commits to.
- **Loosening is compatible, tightening is breaking.** Standard types
  start as loose as possible.
- **Concreteness.** The renderer validates only `config`
  (`pkg/model/component.go`) and each resource
  (`pkg/cmd/template/run.go`) for concreteness. New component-level
  fields can stay non-concrete, and that must remain true.
- **No hardcoded matching.** Templates must be able to accept traits
  that target an interface kind, not only their own GVK.

## Tentative direction

```cue
// What a trait sees of, and what an interface promises about, a component.
#Interface: {
    apiVersion: string
    kind:       string
    // Extension points every implementing template must expose.
    extension?: {...}
    // Config fields every implementing template must accept.
    config?: {...}
}

#Trait: {
    // The kind or interface this trait applies to. When applied, the
    // target component's `metadata` is filled in here, e.g. for resource
    // naming.
    target: {
        apiVersion: string
        kind:       string
        metadata:   _#ObjectMeta
        ...
    }
    // Contributions unified into the target component. References
    // written relative to `apply` (e.g. `config.port`) resolve against
    // the target once embedded, so the trait reads and writes config
    // through the same field. Closed: a trait cannot set `traits`, `ref`,
    // `implements`, `apiVersion` or `kind`.
    apply: {
        config?:    {...}
        resources?: [string]: {...}
        extension?: {...}
    }
    // Trait parameters, e.g. `#mount & {path: "/config"}`.
    ...
}

#ComponentBase: C={
    ...
    // A handle other components use to refer to this one.
    ref?: _
    // Interfaces this component implements, beyond its own kind.
    implements?: [...#Interface]
    // Extension points. The shape is defined by the template.
    extension?: {...}
    // Traits applied to this component.
    traits?: [...#Trait & {
        target: <own GVK, or the GVK of any entry in implements>
        target: metadata: C.metadata
    }]
    for t in traits { t.apply }
    for i in implements {
        if i.extension != _|_ { extension: i.extension }
        if i.config != _|_ { config: i.config }
    }
}
```

The exact CUE expressions (building the GVK disjunction from the
component's own kind and the optional `implements` list, and the
interface conformance comprehension) are left to the implementing change
and must be checked against the CUE version Odin pins.

### Field roles

| Surface           | Kind       | Written by            | Read by                    |
| ----------------- | ---------- | --------------------- | -------------------------- |
| `metadata`        | field      | bundle                | anyone, including traits   |
| `config`          | field      | bundle, traits        | template, anyone           |
| `resources`       | field      | template, traits      | Odin (rendering)           |
| `ref`             | field      | template              | other components           |
| `implements`      | field      | template              | Odin (matching, conformance) |
| `extension`       | field      | traits, bundle        | template                   |
| `traits`          | field      | bundle                | Odin schema (applies them) |
| `#…` declarations | definition | template              | other components (trait constructors, extra handles) |

### What traits can do

A trait can:

- **Read its target's metadata** through `target.metadata`, filled in
  when the trait is applied. With 0002 and 0003 this includes `instance`
  and `namespace`, so traits can name and place the resources they add.
- **Read and write its target's config** through `apply`. Embedding
  `t.apply` into the component re-resolves references inside `apply`
  against the component, so a trait declares the fields it uses and
  refers to them relative to `apply`:

  ```cue
  apply: {
      config: port: int
      resources: svc: spec: ports: [{port: config.port}]
  }
  ```

  `config.port` is the component's port, and anything the trait sets in
  `config` flows into the component. A reference written from the trait
  root (`apply.config.port`) stays bound to the trait's own copy and
  never sees the component's value, so trait authors must use the
  relative form. For a trait targeting an
  interface, the interface's declared `config` shape is what the trait
  can rely on.
- **Contribute to `resources` and `extension`** through `apply`. Adding
  a resource (an HTTPRoute, a PodDisruptionBudget) or setting config
  needs no cooperation from the template.

Two consequences of CUE's evaluation model are worth stating for trait
authors:

- **Traits see the final config.** Unification has no order, so
  a trait sees the component's config after every trait's contributions,
  its own included. There is no "config before traits" view.
- **A trait must not derive a config field from itself.** Reading
  `config.port` to build a resource is fine. Defining a config field in
  terms of itself is caught by Odin's existing config validation, but
  reported as a non-concrete value (`config.m: non-concrete value int in
  operand to +`) at the trait's source line, not as a cycle. If nothing
  reads the field, it is still caught there, since every component's
  `config` is validated.
- **The trait values in `traits` stay incomplete.** Only the embedded
  copy of `apply` sees the component. The list elements themselves keep
  unresolved fields like `config.port: int`. Odin is unaffected, because
  it only validates `config` and exports `resources`. But `cue vet -c` or
  `cue export` over a whole bundle fails on them, and tools that export
  whole components (schema export, the YAML tooling in 0008) must skip
  `traits` when checking concreteness.

These semantics were checked in a scratch prototype against CUE v0.17.1:
with references relative to `apply` and a plain
`for t in traits { t.apply }`, a trait reads the component's config and
its writes override template defaults (`*1 | int` → `3`); references
from the trait root get an incomplete value; whole-component export fails
only on the `traits` list.

`apply` is closed, so a trait cannot set `traits`, `ref`, `implements`,
`apiVersion` or `kind`. That turns what would otherwise be a cycle or
conflict, which CUE reports badly, into a plain "field not allowed"
error at the trait.

### `extension`

The template's public extension points. The workload template would
declare:

```cue
extension: {
    containers: [string]: _#ContainerConfig
    volumes: [Name=string]: schema.#VolumeSpec & {name: Name}
}
```

and read `extension.containers` / `extension.volumes` when rendering.
Because a template definition closes its `extension` struct, a trait
writing `apply: extension: contaners: ...` is rejected.

Bundle authors may also set `extension` directly, without a trait.

### Interfaces and `implements`

An interface names a kind that several templates share and declares what
those templates must expose. The workload module would define:

```cue
#WorkloadInterface: odin.#Interface & {
    apiVersion: "platform.vituity.com/v1alpha1"
    kind:       "Workload"
    extension: {
        containers: [string]: _#ContainerConfig
        volumes: [string]: schema.#VolumeSpec
    }
}

_#Workload: C=odin.#ComponentBase & {
    implements: [#WorkloadInterface]
    extension: { containers: ..., volumes: ... }
    // _allowedTraits, the #traits typing and the for-loop are removed.
    ...
}
```

Two rules follow:

- **Matching.** A trait applies to a component if its `target` GVK equals
  the component's own `apiVersion`/`kind`, or the GVK of any interface in
  `implements`.
- **Conformance.** Each implemented interface's `extension` and `config`
  are unified into the component. A template that claims to implement
  `Workload` but doesn't expose `extension.containers` fails because its
  closed `extension` struct rejects the field. So a trait written against
  an interface is guaranteed the extension points it relies on.

`target` and `implements` are concrete data, so Odin can index traits by
GVK and answer "which traits fit this component?" without evaluating CUE.
It can also report mismatches precisely: "trait `#env` targets
ConfigMap; `myapp` (Deployment) accepts Deployment, Workload."

`_#Trait` can stay in the workload module as a convenience for trait
authors that presets `target` to the Workload GVK.

### `ref`

Set by the template, read by bundles. One primary handle per component.
A component that needs to expose several things uses a struct. Extra
handles can still be published as `@odin(ref)` declarations.

## Options considered

### Option A — Keep the convention

- **Con:** Silent typos, invisible matching, no cross-module contract.
  Rejected.

### Option B — Standardize as optional definitions (`#ref?`, `#traits?`, …)

- **Con:** Keeps the closedness workaround, and with it the silent-typo
  problem, even though the schema now knows the fields. Rejected.

### Option C — Hardcoded GVK matching

`traits` accepts only traits targeting the component's own GVK.

- **Con:** Breaks interface kinds such as `Workload`. Rejected.

### Option D — Minimal: `traits` plus application, matching left to templates

`#ComponentBase` declares `traits?: [...#Trait]` and applies them.
Templates narrow the element type themselves, as `_allowedTraits` does
today.

- **Pro:** Smaller schema. Odin has no matching rule at all.
- **Con:** Matching stays invisible to tooling, and a template that
  doesn't narrow accepts any trait.

### Option E — Traits write only into `extension`

A trait's contribution is unified into the component's `extension` and
nowhere else. Templates that want traits to add resources expose that
explicitly.

- **Pro:** The tightest contract between traits and templates.
- **Con:** Traits that add a resource or set config need every template
  to cooperate. Rejected: traits must always be able to reach `config`
  and `resources`.

### Option F — Traits as whole partial components

Keep today's semantics: a trait's extension is a partial
`#ComponentBase` unified into the whole component.

- **Con:** Nothing stops a trait from setting `traits`, `apiVersion` or
  `kind`, which produces cycles and conflicts that CUE reports badly. The
  closed `apply` block keeps the same reach into `config` and
  `resources` without that hazard.

### Option G — Interface inheritance

Let an `#Interface` extend other interfaces, with matching and
conformance following the hierarchy.

- **Deferred.** Transitive matching makes the static trait index harder
  to reason about, diamond-shaped hierarchies raise conformance questions
  CUE won't answer clearly, and error messages would need to explain the
  chain. A template listing several interfaces in `implements` covers
  the known cases with flat composition. If inheritance turns out to be
  needed it can be added later without breaking anything.

## Migration (contributes to 0006)

Platform modules:

- `#ref` → `ref`.
- `#containers`, `#volumes` and similar extension points → fields under
  `extension`.
- `#traits` typing and application → removed; declare `implements`
  instead.
- Interface kinds such as `Workload` → an `odin.#Interface` value
  declaring the shared extension points.
- `core.Trait` → `odin.#Trait`. `#target` → `target` (metadata is now
  filled in by the schema). `#extension` → `apply`, with extension-point
  writes moved under `apply: extension` (for example
  `#containers: workload: env: [...]` becomes
  `apply: extension: containers: workload: env: [...]`) and config and
  resource writes under `apply: config` / `apply: resources`.
- Trait constructors (`#env`, `#mount`, …) are unchanged apart from the
  above.

Bundle authors writing CUE: `mygateway.#ref` → `mygateway.ref`,
`#traits: [...]` → `traits: [...]`.

## Open questions

- **Trait access to `metadata`.** Traits can read the target's metadata
  but not write it. Is there a need for traits to add labels or
  annotations at the component level, beyond what `config` exposes?
- **Sub-bundle handles.** Should a `#BundleRef` (0001) be able to expose
  a `ref`, so components can attach to a composed sub-bundle?
- **Multiple handles.** Is one `ref` plus `@odin(ref)` declarations
  enough?

## Consequences

To be populated once the open questions are settled.

## References

- [0001 — Bundle composition](0001-bundle-composition.md)
- [0006 — `v1alpha2` schema cohort](0006-v1alpha2-cohort.md)
- [0008 — YAML bundle format](0008-yaml-bundle-format.md)
- `pkg/schema` — `@odin(...)` declaration categories
