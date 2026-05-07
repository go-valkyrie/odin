# 0001 — Bundle composition

- **Status:** Proposed (2026-04-23)
- **Deciders:** Odin maintainers
- **Consulted:** n/a

> Part of the [`v1alpha2` schema cohort](0006-v1alpha2-cohort.md) — the
> trigger ADR. Cohort-wide concerns (release expectations, migration,
> GVK-gate rationale) live in 0006.

> This ADR is an active working document. Multiple issues and options remain
> open. It will move to Accepted once a direction is chosen and the
> consequences are understood; until then, sections below may be revised.

## Context

An Odin bundle today is a flat collection of components: the bundle file
(typically `bundle.cue`) unifies `odin.#Bundle` with a `metadata` block, a
`components` map whose entries are instances of component templates, and a
`values` block supplying configuration. A component template is a CUE
definition (living in a CUE module, conventionally authored by platform
engineers) that unifies with `#ComponentBase` and produces Kubernetes
resources. Bundles are distributed as OCI artifacts via `odin push` / `odin
pull`, and component-template CUE modules are distributed through the
standard CUE module registry mechanism.

This model has a clean separation of concerns: platform engineers publish
reusable primitives (component templates), developers compose them into
application-shaped bundles. Reuse within that model is straightforward.

What the model does not support is **bundle-level reuse** — one developer
authoring a bundle that builds on another developer's bundle. The scenario
is: team A publishes a bundle for "our standard web application" (deployment,
service, ingress, a secret, baseline configuration). Team B wants the same
thing plus a CronJob. Today team B's options are:

1. Copy team A's bundle and diverge. Loses sync, duplicates effort.
2. Ask team A to decompose the bundle into component templates that team B
   reassembles. This works, but pushes developers into platform-engineer
   authorship — writing CUE definitions that unify with `#ComponentBase`,
   publishing CUE modules, managing a separate distribution channel. That is
   a deliberately different authoring persona, and forcing it on every shared
   application pattern collapses the persona distinction that makes the
   current model work.
3. Wait for a first-class composition mechanism. (This ADR.)

## Problem statement

Odin needs a way for one bundle to include another bundle, with value
configuration and bounded customization, without requiring the sub-bundle to
be rewritten as a set of component templates.

## Non-goals

- Runtime composition (stitching rendered YAML from independently-rendered
  bundles). Odin is a build-time tool; composition should happen during CUE
  evaluation.
- Kustomize-style post-render patching. If users want that, they can run
  Kustomize on Odin's output. We are not replicating it inside the tool.
- Replacing component templates. Component templates remain the platform-
  engineer primitive. Bundle composition is additive.

## Forces and constraints

- **Pure CUE.** Composition should be expressed in the CUE evaluation model,
  not as a preprocessor or a separate overlay language.
- **Backward compatibility.** Existing flat bundles must continue to load,
  render, and publish unchanged.
- **Developer-authorable.** The sub-bundle mechanism must be usable by the
  same persona that writes bundles today. If composing a bundle requires the
  skills or tooling of publishing a component-template module, we have
  recreated the problem we are trying to solve.
- **Compatible with existing distribution.** Whatever we choose must have a
  coherent story for how sub-bundles are fetched, versioned, and cached.

## Known issues to resolve

These are the concrete obstacles to any composition design. Each option in
the next section is partly characterized by how it answers these.

1. **OCI artifact format incompatibility.** The divergence between Odin's
   bundle artifacts and CUE's module artifacts is structural, not just a
   media type label:
   - Odin pushes standard ORAS-style tarball layers under the artifact
     type `application/vnd.odin.bundle.v1`.
   - CUE publishes module artifacts as `.zip` layers and attaches
     additional module metadata (notably `cue.mod/module.cue` as its own
     discrete layer).

   So `cue mod tidy` cannot consume an Odin-published bundle even if the
   layer contents are otherwise a valid CUE module. Possible responses:
   - **Migrate.** Change `odin push` to produce CUE-module-shaped
     artifacts. This is more than a media type change — it is a new
     packaging pipeline (zip, metadata layer). Existing consumers of the
     old artifact type need a compatibility path.
   - **Dual-publish.** `odin push` produces both artifact types so CUE
     tooling and existing Odin consumers both work. Costly because the
     two formats are genuinely different artifacts, not different manifests
     over the same blob.
   - **Custom CUE module resolver.** Odin implements a resolver against
     `cuelang.org/go/mod/modregistry` that pulls Odin-format bundles
     through the existing pull path, extracts them, and presents the
     extracted tree to CUE's module graph as if it were a CUE module
     artifact. Bundle artifact format is unchanged. The resolver is the
     translation layer. Also requires a mapping in `odin.toml`, since
     bundle OCI repositories do not follow CUE's package-name-to-registry
     conventions.
   - **Skip CUE module resolution entirely.** Composition happens outside
     the CUE module system (Option 4).

   Each choice interacts with the source-mechanism sub-decision below.

2. **Component name collisions.** When parent and sub-bundle both use the
   same component name, CUE will unify them — which is a bug disguised as a
   feature. We need a collision policy: error, scope, or rename.

3. **Values schema merging.** Odin's values schema is derived from the
   bundle's components. With composition, the parent's values schema must
   grow to include the sub-bundle's, in a form that `odin show values`
   produces sensibly.

4. **Override semantics.** When a parent wants to modify a sub-bundle's
   component, does it go through values only, or can it unify arbitrary CUE?
   Values-only preserves encapsulation; unification is pure CUE but leaky.

5. **Attribution at render time.** Does the tooling need to answer "which
   resources came from which sub-bundle?" If yes, the composition model must
   preserve that information through evaluation. If no, we can dissolve
   sub-bundle identity once it's merged.

6. **Recursive composition.** Can a sub-bundle itself have sub-bundles? If
   yes, resolution and caching need to handle the graph. If no, we have to
   explain a surprising restriction.

## Options considered

### Option 1 — Flat merge via an `imports` field

The parent declares sub-bundles via a new `imports` field; the sub-bundle's
`components` and `values` are unified into the parent's at the same level.
Render output is indistinguishable from a hand-written flat bundle.

```cue
odin.#Bundle & {
    metadata: name: "my-variant"
    imports: [webappbase.#Bundle]
    components: cronjob: workload.#CronJob & { ... }
    values: components: myapp: replicas: 3
}
```

**Pros**

- Smallest conceptual addition: one new top-level field.
- Values schema extension is automatic via CUE unification.
- Overrides are natural CUE: unify onto `components.myapp` to modify it.
- Rendered output stays flat; existing tooling works unchanged.
- Recursive composition composes for free (sub-bundle's own imports resolve
  during its load).

**Cons**

- Name collisions between parent and sub-bundle surface as CUE unification
  errors. Author must rename, and cannot rename a sub-bundle's components
  without reaching into the sub-bundle's scope.
- Sub-bundle identity dissolves after merge. Attribution requires preserving
  the `imports` declaration through evaluation and a tooling path to query
  it.

### Option 2 — Namespaced sub-bundles

A new top-level field (e.g. `subBundles`) creates a named scope. The sub-
bundle's components live under `subBundles.<name>.components`, values under
`values.subBundles.<name>.components`. Render output either retains the scope
as a prefix (`<name>.<component>.<resource>`) or flattens via policy.

```cue
odin.#Bundle & {
    metadata: name: "my-variant"
    subBundles: base: webappbase.#Bundle
    components: cronjob: workload.#CronJob & { ... }
    values: subBundles: base: components: myapp: replicas: 3
}
```

**Pros**

- No collisions: parent and each sub-bundle have separate component
  namespaces.
- Sub-bundle identity is preserved through evaluation; attribution is
  trivial.
- Composition becomes a first-class concept, visible in the CUE tree.

**Cons**

- Introduces a second top-level shape alongside `components`. Authors must
  learn when to reach for which.
- Values tree gains a new level of nesting; `odin show values` output
  becomes more complex.
- Overrides into sub-bundle components go through a longer path
  (`subBundles.base.components.myapp.<field>`), which is more verbose than
  unifying directly onto a component.
- Requires a rendering policy decision (keep or flatten the prefix in
  resource output).

### Option 3 — Bundle reference as a component (via a built-in template)

The Odin API gains a `#BundleRef` component template. A sub-bundle occupies
a component slot in the parent; the template flattens the sub-bundle's
components' resources into its own `resources` field via a CUE
comprehension. The referenced bundle value is obtained through a CUE import,
which requires a module resolver (source mechanism (b)) to map module names
to bundle OCI artifacts.

Template shape (lives in the Odin API, authored once):

```cue
#BundleRef: #ComponentBase & {
    apiVersion: "odin.go-valkyrie.com/v1alpha1"
    kind:       "BundleRef"
    bundle: #Bundle
    config: bundle.values
    resources: {
        let configured = bundle & {values: config}
        for compName, comp in configured.components
        for resName, res in comp.resources {
            "\(compName)-\(resName)": res
        }
    }
}
```

The `bundle` reference lives outside `config` because it's a CUE-time
composition input, not runtime configuration — values files (YAML/JSON)
cannot express imported bundle values. `config` is bound to `bundle.values`
so the BundleRef's configuration surface is exactly the sub-bundle's
configuration surface — including any bundle-level fields, not just
per-component configs. Parent overrides arrive via the standard
`config: values.components[Name]` wiring, then the `let` in the resources
comprehension unifies those overrides back into the bundle's values before
iterating, ensuring rendered resources reflect the parent's configuration.

Bundle author usage:

```cue
import "myorg/webappbase"

odin.#Bundle & {
    metadata: name: "my-variant"
    components: {
        base: odin.#BundleRef & {
            metadata: name: "base"
            bundle: webappbase
        }
        cronjob: workload.#CronJob & { ... }
    }
}
```

The values file mirrors the sub-bundle's `config` shape under the parent's
slot:

```yaml
components:
  base:
    components:
      myapp:
        image: custom
        replicas: 3
  cronjob:
    schedule: "0 * * * *"
```

The conceptual framing: a component is a unit of deployment intent that
produces resources. `#WebApp` produces multiple resources because it
represents a logically composite thing. `#BundleRef` extends that same idea
one level deeper — a composite whose resources come from flattening another
bundle's components.

**Pros**

- **Almost no new Odin machinery.** One definition in the Odin API, plus a
  working CUE module resolver. No bundle-schema changes.
- **Composition is expressed in pure CUE.** The flattening comprehension,
  override semantics via unification, and attribution via path all emerge
  from the language; Odin contributes no bespoke merge logic.
- **Values flow through nested YAML.** Because `config: bundle.values`, the
  parent's `values.components.base` flows transitively into the sub-bundle's
  configuration via the existing `config: values.components[Name]` wiring.
  `odin show values` on the parent surfaces the full nested schema for
  free, including bundle-level fields, not just per-component configs.
- **Overrides are native CUE.** Standard values flow handles most cases,
  and direct unification on the imported bundle is available at the import
  site as an escape hatch:
  ```cue
  bundle: webappbase & {
      values: components: myapp: replicas: 3
  }
  ```
- **Natural namespacing.** The component slot name prefixes the flattened
  resources; no collisions between parent components and sub-bundle
  resources at the parent level.
- **Recursive composition is transparent.** A sub-bundle can itself include
  a `#BundleRef` component; the parent's comprehension sees the nested
  flattening without knowing about it.

**Cons**

- **Resource name flattening needs a documented policy.** The template
  above uses `<subcomponent>-<resource>` as the key. Collisions with the
  parent's own resource names are the author's responsibility to avoid.
- **Depends on the module resolver.** Source mechanism (b) is not optional
  for this shape; without it, the CUE import fails.
- **Override mechanics depend on how defaults are exposed.** Out of the box,
  CUE unification narrows rather than overwrites: a sub-bundle that sets
  concrete values (e.g. `image: "nginx:latest"`) cannot be overridden by a
  parent without producing a CUE conflict. Sub-bundle authors must use CUE
  defaults (`image: string | *"nginx:latest"`) for any field parents may
  override, or we need an out-of-band defaults mechanism. This applies to
  any composition shape but is most visible in Option 3 because the
  override path is shortest.

### Option 4 — Odin-orchestrated children (loader-side composition)

Composition is not a CUE-level concern. `odin.#Bundle` gains a `children`
field that takes OCI references. Odin's Go loader fetches each child
(reusing the existing `odin pull` path), loads it as a Bundle, and composes
its components and values into the parent tree before handing the result to
the CUE evaluator. The CUE file declares *what* to compose; the Go code
decides *how*.

```cue
odin.#Bundle & {
    metadata: name: "my-variant"
    children: [
        {from: "oci://registry/webappbase:v1"},
    ]
    components: cronjob: workload.#CronJob & { ... }
    values: components: myapp: replicas: 3
}
```

This is not another point on the same axis as Options 1-3; it reframes the
problem. The merge shape question (flat vs. namespaced vs. slotted) still
applies within this option, but the loader implements whichever shape we
pick rather than expressing it through CUE constructs.

**Pros**

- Reuses the existing OCI bundle pull path. No CUE module registry
  dependency, no new artifact format. Known Issue #1 (OCI artifact type
  mismatch) is no longer on the critical path for composition.
- Composition semantics are ours to define: collision handling,
  attribution, version pinning, merge order, diagnostics. We don't have to
  fight CUE's evaluation model to get the behavior we want.
- The existing bundle cache handles sub-bundle caching; no coordination
  with CUE's module cache.
- Attribution is trivial. The loader knows the provenance of every
  component and can expose it to tooling.
- Recursive composition is a loop in Go rather than a CUE module graph.
- `odin push` and `odin pull` are unchanged.

**Cons**

- Composition is invisible to raw CUE tooling. `cue eval` on a bundle file
  no longer produces the composed bundle — the file is no longer a
  complete description of its value. Anyone debugging with the `cue` CLI
  directly will see a partial picture.
- Odin owns more semantics, more code, and more documentation burden.
  Collision policy, override semantics, and merge ordering all become our
  design and maintenance responsibility.
- Error messages during composition may reference content the user did not
  author (a child bundle's component), which is harder to diagnose than a
  CUE unification error sourced from the user's own file.
- The merge-shape question is relocated, not answered. Option 4 says
  "Odin performs the merge," not "Odin performs *this specific* merge."
- Values override design must be specified from scratch: merge order
  between parent and child values, whether children carry their own values
  files, what is overridable from where.

## Cross-cutting sub-decisions

These apply to the CUE-native options (1-3). Option 4 collapses the source
mechanism and OCI artifact sub-decisions by definition (Odin-native fetch,
existing artifact type); only override semantics remains open for it.

### Source mechanism

How does a parent find a sub-bundle? All variants below keep composition
itself CUE-native (the parent's CUE file expresses composition through
imports and unification); they differ in how the import resolves to bytes.

- **(a) Standard CUE module resolution with migrated artifact.** Bundles
  are published as CUE module artifacts — zip layers with the required
  module-metadata layer (see Known Issue #1). `cue mod tidy` and CUE's
  built-in registry support work unchanged. Lowest implementation cost on
  the Odin side post-migration; cost is paid up front in `odin push`
  rework and consumer-compatibility handling.
- **(b) Custom CUE module resolver.** Odin implements a resolver against
  `cuelang.org/go/mod/modregistry` that maps module names to bundle OCI
  artifacts using an `odin.toml` mapping, pulls them through the existing
  bundle pull path, and translates the odin-format tarball into the shape
  CUE's module system expects before handing it to the evaluator. Bundle
  artifact format is unchanged. Preserves the CUE-native authoring
  experience (`import`, unification) at the cost of Odin owning a
  module-resolution implementation, a format-translation step, and an
  additional configuration surface. Tools that invoke CUE directly without
  going through Odin (e.g. `cue eval` on a raw file) will not see
  bundle-packaged dependencies unless the resolver is also made available
  to them.
- **(c) Dual-publish.** Publish both artifact shapes from `odin push` —
  the existing ORAS tarball plus a CUE-module-shaped zip with the metadata
  layer. Lets (a) work for new bundles while older consumers keep using
  the original format. Highest operational cost because the two artifacts
  are genuinely distinct, not two manifests over shared blobs.

### Override semantics

How does a parent modify a sub-bundle's components?

- **Values-only.** The sub-bundle author decides what is configurable by
  designing their values schema. Clean encapsulation, real limits on what
  downstream can do.
- **Values-primary with CUE unification escape hatch.** Values are the
  documented interface; unifying onto a sub-bundle's component is always
  possible when values are insufficient. Matches how component templates
  work today.
- **Explicit patch blocks.** A separate `overrides` or `patches` field
  where parent expresses modifications. Adds surface area; arguably
  redundant with CUE unification.

### OCI artifact strategy

Decided in concert with the source mechanism.

- **(i) Keep the current ORAS tarball format.** Compatible with source
  mechanism (b) (custom resolver) and with Option 4 (loader-orchestrated
  children). No migration cost.
- **(ii) Migrate to the CUE module OCI artifact format.** Required for
  source mechanism (a). Means changing the `odin push` packaging pipeline
  to produce zip layers plus the module-metadata layer, not just relabeling
  the media type. Breaks existing `odin pull` against registries serving
  the old format unless a compatibility path is retained.
- **(iii) Dual-publish.** Publish both artifact formats from `odin push`.
  Enables (a) for new consumers while keeping (i)'s compatibility
  guarantee. Highest operational cost — two distinct packaging pipelines,
  two distinct artifacts in the registry.

## Tentative direction

**Current lean: Option 3 (bundle reference as a component via `#BundleRef`)
+ source mechanism (b) (custom CUE module resolver) + values-primary with
CUE unification escape hatch.**

The Option 3 re-framing as a built-in component template — rather than a
bundle-schema concept — collapses the surface area substantially. A single
definition in the Odin API does the heavy lifting through a CUE
comprehension, and overrides, attribution, and recursive composition all
fall out of native CUE semantics. The remaining scope is the custom
resolver, its `odin.toml` mapping surface, and the translation from Odin's
ORAS tarball format into the shape CUE's module system expects.

Held against the other paths:

- **vs. Option 1 (flat merge).** Option 1 has the appealing property that
  rendered output looks like a hand-written flat bundle, but it requires a
  new `imports` field on the bundle schema and leaves name-collision
  management to authors with no scope-level tools. Option 3 keeps
  everything inside `components` and gets namespacing for free via the
  slot name.
- **vs. Option 2 (namespaced sub-bundles).** Option 2 achieves the same
  isolation as Option 3 but through a parallel top-level structure, which
  duplicates machinery and complicates the values tree. Option 3 is
  strictly smaller.
- **vs. Option 4 (Odin-orchestrated children).** The two paths produce
  surprisingly similar user-visible properties: both keep the bundle
  artifact format, both require Odin in the loop to evaluate a composed
  bundle. The distinction is where composition *logic* lives — CUE
  unification (Option 3) or Go code (Option 4). Option 3 wins on
  consistency with how the rest of Odin's composition machinery already
  works (component templates are CUE-native); Option 4 wins on implementation
  simplicity since it avoids the CUE module resolver and translation layer.

If the custom resolver turns out to be materially harder than expected, or
if the format-translation layer raises fidelity issues we can't resolve,
**Option 4 is the natural fallback**: same user-visible shape (component
slot, OCI reference, Odin resolves), composition logic in Go instead of
CUE.

## Open questions

- How is a bundle exposed for import? A bundle's package is already
  importable — its top-level value is the bundle — so `import "foo/bar"`
  followed by referring to the package identifier works out of the box.
  Do we bless that as the convention, additionally encourage a named
  definition (e.g. `#Bundle`) for clarity, or require one? What, if
  anything, does Odin check at import time?
- Is recursive composition supported in v1? If yes, what cycle detection do
  we need? If no, how do we explain the restriction?
- Does attribution (resource → source sub-bundle) need a tooling path, and
  if so, what is it?
- How does `odin show values` render a composed schema — flat, grouped by
  sub-bundle, or both views available?
- How do values files at the filesystem level interact with sub-bundle
  values? Does the parent's `values.yaml` reach into sub-bundle values, or
  does each bundle bring its own defaults?
- Does the migration path for existing bundles require a format bump, and if
  so, how does the `compat` policy handle it?

## Consequences

To be populated once a direction is settled.

## References

- Existing bundle schema: `api/v1alpha1/bundle.cue`,
  `api/v1alpha1/component.cue`, `internal/schema/bundle.cue`
- Bundle loader: `pkg/model/bundle.go`
- Component template discovery: `pkg/model/componenttemplate.go`
- OCI artifact type: `pkg/oci/oci.go` (`application/vnd.odin.bundle.v1`)
