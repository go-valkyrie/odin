# 0008 — YAML bundle format

- **Status:** Proposed (2026-11-05)
- **Deciders:** Odin maintainers
- **Consulted:** n/a

> Depends on the [`v1alpha2` schema cohort](0006-v1alpha2-cohort.md),
> in particular [0007](0007-component-handles-and-traits.md). This ADR
> is not itself part of the cohort: it changes the loader, not the
> schema, and YAML bundles are `v1alpha2`-only.

## Context

Odin distinguishes two authoring personas. Platform engineers write
component templates in CUE. Developers compose those templates into
bundles. ADR 0001 lists "developer-authorable" as a constraint.

In practice, a bundle file is almost entirely data:

```cue
odin.#Bundle & {
    metadata: name: "webapp-basic"
    components: myapp: workload.#WebApp & { metadata: name: "myapp" }
    values: components: myapp: { image: "nginx:latest", replicas: 1 }
}
```

The only parts that are not plain data are references: to a template
(`workload.#WebApp`), and between components (`mygateway.ref`,
`configmap.#mount & {path: "/config"}`).

We want GUI tools in which developers author bundles. A GUI needs a
format it can read and write losslessly. Editing arbitrary CUE source
is lossy: any expression the GUI does not model has to be either
preserved verbatim or dropped. A constrained YAML format solves that,
as long as it can express the references bundles actually use.

## Problem statement

Define a YAML representation of `v1alpha2` bundles that:

- expresses everything a typical bundle file does, including imports,
  template references, sub-bundle composition and wiring between
  components;
- can be analyzed without evaluating CUE, so a GUI can show the
  reference graph, list valid targets and detect cycles;
- adds no semantics beyond CUE's own, and none beyond the standard
  fields of 0007;
- compiles to CUE, which remains the only evaluation engine.

## Non-goals

- **YAML component templates.** Templates remain CUE.
- **A general expression language in YAML.** No arithmetic,
  conditionals or comprehensions. The only computation is calling a
  function with literal arguments (see *Path grammar*). Bundles that need them use CUE (see
  *Mixing YAML and CUE*).
- **YAML values files.** These are already supported.
- **`v1alpha1` support.**

## Forces and constraints

- **CUE is the engine.** YAML bundles must evaluate, validate and
  render through exactly the same path as CUE bundles, so the two can
  never behave differently.
- **Static analyzability.** Every reference must be resolvable to a
  schema location without evaluation.
- **Genericity.** Odin stays agnostic about platform patterns (traits,
  handles beyond 0007, constructor definitions). The YAML format must
  not encode any of them.
- **GitOps hygiene.** Bundle files pass through CI templating
  (`envsubst`, Helm-style pipelines). Syntax like `${...}` would be
  replaced before Odin sees it.
- **Errors point at YAML.** A type mismatch must be reported at
  `bundle.yaml:14:20`, not at a position in generated CUE.

## Options considered

### Option A — GUI edits CUE directly

No new format. Tools read and write `bundle.cue` via `cue/ast`.

- **Con:** Lossy round-trip for any construct the tool doesn't model.
  Every tool reimplements CUE source manipulation.

### Option B — YAML with string interpolation (`"${components.db.ref}"`)

- **Con:** Ambiguous with literal strings, needs escaping rules, and
  collides with CI templating. Rejected.

### Option C — YAML with `{$ref: "..."}` objects

- **Pro:** Representable in JSON.
- **Con:** Verbose, and collides with real `$ref` keys in values (for
  example a ConfigMap embedding a JSON Schema). Kept only as the JSON
  form of option D (see *JSON equivalence*).

### Option D — YAML with a small set of custom tags

`!ref` and `!fmt` for references, plus `!type` and `!enum` for schema,
each restricted to paths, constants and plain data.

- **Pro:** Unambiguous, keeps positions, survives CI templating, and is
  a distinct node type for editors. Supported by `yaml.v3` and the
  common JavaScript YAML libraries.
- **Con:** Tags are less familiar to some users, and JSON cannot
  represent them directly.

### Option E — Option D plus `!cue <expr>`

An unrestricted fallback for arbitrary CUE expressions.

- **Con:** Breaks static analyzability for the whole file. *Mixing YAML
  and CUE* covers the same need without that cost. Rejected.

## Tentative direction

**Option D.**

### File shape

```yaml
apiVersion: odin.go-valkyrie.com/v1alpha2
kind: Bundle
metadata:
  name: shop
imports:
  gateway:    example.com/platform/gateway
  cfg:        example.com/platform/config
  workload:   example.com/platform/workload
  security:   example.com/platform/security
  schema:     example.com/platform/schema
  webappbase: example.com/bundles/webappbase
components:
  mygateway:
    template: gateway.#Gateway
  configmap:
    template: cfg.#ConfigMap
    config:
      data: { LOG_LEVEL: info }
  myapp:
    template: workload.#Deployment
    traits:
      - !ref components.configmap.#env
      - !ref
        path: components.configmap.#mount
        with:
          path: /config
      - !ref security.#SecureDefaults
  base:
    bundle: webappbase                            # sub-bundle (0001)
  myroute:
    template: gateway.#HTTPRoute
    config:
      gateway: !ref components.mygateway          # short for .ref
      backendPort: !ref components.myapp.config.port
      hostname: !fmt "{metadata.name}.{values.domain}"
values:
  # Base domain for all routes.
  domain!: !type string
  components:
    myapp:
      image!: !type string
      port: 8080
      hostname: !fmt "shop.{values.domain}"
      replicas?: !type uint & >0
      logLevel: !enum [debug, info, warn]
      imagePullPolicy?: !type schema.#ImagePullPolicy
defaults:
  components:
    myapp:
      logLevel: info
```

Top-level keys follow `#Bundle` — `apiVersion`, `kind`, `metadata`,
`components`, `values`, and `defaults` (0005) — plus `imports`, which
has no counterpart in `#Bundle` and becomes the generated file's import
declarations.

A component accepts exactly these keys; anything else is an error:

| Key        | Maps to                                       |
| ---------- | --------------------------------------------- |
| `template` | the template definition the component unifies with: a path rooted at an import (`workload.#Deployment`) |
| `bundle`   | a sub-bundle, in place of `template`: a path rooted at an import (`webappbase`). Compiles to `odin.#BundleRef & {bundle: …}` (0001) |
| `metadata` | `metadata`                                    |
| `config`   | `config`                                      |
| `traits`   | `traits` (0007)                               |
| `extension` | `extension` (0007) — set extension points directly, without a trait |

`ref` and `implements` are set by templates, never by bundles, so they
are not component keys.

There is no way to set arbitrary definition fields from YAML. Writing
into definitions is the silent-typo pattern 0007 removes, and 0007's
standard fields cover what bundles set. Anything else goes in a `.cue`
file alongside the YAML.

### Imports

```yaml
imports:
  workload:   example.com/platform/workload
  workloadv2: example.com/platform/workload@v2          # major version
  legacy:     example.com/platform/workload@v1:workload # major + package qualifier
```

`imports` maps an alias to a CUE import path: a package path with an
optional `@v<major>` version and an optional `:package` qualifier,
exactly as CUE spells it in an `import` declaration. Every reference into another package goes through
an alias: templates (`template: workload.#Deployment`), traits and other
definitions (`!ref security.#SecureDefaults`), named types
(`!type schema.#ImagePullPolicy`) and sub-bundles (`bundle: webappbase`).
There is no inline form such as `example.com/platform/workload#WebApp`.
One spelling keeps the file easy to round-trip, and the import list is
where a GUI shows and edits what a bundle depends on.

- **Aliases are identifiers.** `components`, `values`, `metadata`,
  `defaults` and `odin` are reserved, because they are path roots or
  already imported.
- **`odin` is always imported.** The generated file needs the api
  package anyway, so `!type odin.#Name` and similar work without
  declaring it.
- **Unused imports are an error**, reported at the YAML line, matching
  CUE's own rule.
- **Major versions are optional.** A bundle can depend on several major
  versions of the same module at once, under different aliases, for
  example while migrating components from one platform generation to
  the next. Only the major version goes in the import path; the exact
  version stays in `cue.mod/module.cue`. When `module.cue` depends on
  more than one major version of a module, every import of that module
  must name its major. Odin reports a missing one at the YAML line
  rather than as CUE's "cannot find module providing package".

  Checked on CUE v0.18.0-rc.3 against a local registry: importing
  `example.com/plat/workload@v0` and `example.com/plat/workload@v1` in
  one package works, and `cue mod tidy` records both
  `example.com/plat@v0` and `example.com/plat@v1`. A full version in an
  import path (`@v1.0.0`) is rejected with "import paths can only contain
  a major version specifier", and an unversioned import fails once both
  majors are present.
- **Paths rooted at an alias** may select any non-private value or
  definition in the imported package. Static validation loads the
  imported packages to check them, but does not evaluate the bundle.

#### Sub-bundles

A component with `bundle: <path>` instead of `template:` composes
another bundle via 0001's `#BundleRef`:

```yaml
components:
  base:
    bundle: webappbase
values:
  components:
    base:
      components:
        myapp: { replicas: 3 }
```

The path is rooted at an import alias. Whether it names the package
itself (`webappbase`) or a definition in it (`webappbase.#Bundle`)
follows whatever 0001 settles as the convention for exposing a bundle.
Values reach the sub-bundle through the parent's `values`, exactly as in
0001. `metadata`, `config` and `traits` keys work as for any other
component.

#### YAML and CUE bundles import each other

A YAML bundle compiles to an ordinary CUE package whose top-level value
is the bundle, the same shape 0001 assumes for CUE bundles. So a CUE
bundle can import a YAML bundle, a YAML bundle can import a CUE bundle,
and YAML bundles can import each other, all with the same `imports`
mechanism.

This needs the YAML to be compiled wherever a bundle package is loaded,
not only for the top-level bundle:

- **Fetched bundles** are compiled by the module resolver from 0001
  (source mechanism (b)), which already translates bundle artifacts into
  the shape CUE expects. Compiling `bundle.yaml` becomes part of that
  translation.
- **Local and vendored packages** are compiled by the loader and passed
  to CUE as overlay files.

Template modules are unaffected: they are plain CUE modules.

Module versions stay in `cue.mod/module.cue`, managed by
`odin cue mod tidy` / `get`. See *Open questions*.

### Tags

**`!ref <path>`** — a reference. In scalar form it compiles to the
referenced value. In mapping form, it combines one reference with one
block of data:

```yaml
!ref
path: components.configmap.#mount
with: { path: /config }
```

This compiles to `<ref> & {path: "/config"}`. `with` is plain YAML and
may itself contain `!ref` and `!fmt`. Closed definitions reject unknown
`with` keys, so typos are caught.

A `!ref` path can start at an import alias as well as at `components`,
`values` or `metadata`. That covers traits and other definitions that
don't derive from another component, such as a standard
security-defaults trait: `!ref security.#SecureDefaults`, or with
parameters:

```yaml
!ref
path: security.#SecureDefaults
with: { runAsUser: 1000 }
```

**`!fmt "<text with {path} segments>"`** — string interpolation. Each
`{…}` contains a path under the `!ref` grammar and nothing else. `{{`
and `}}` produce literal braces.

### Values schema

A bundle's `values` block is a schema as much as it is data: it pins
some values, constrains others, and declares the global values
deployers must supply. YAML scalars only express concrete values, so
the format adds a restricted way to write constraints.

**`!type <expr>`** — a CUE constraint expression from a restricted
subset:

| Allowed                     | Example                          |
| --------------------------- | -------------------------------- |
| Basic types                 | `string`, `int`, `uint`, `float`, `number`, `bool`, `bytes`, `_` |
| Bounds                      | `>0`, `>=1 & <=10`               |
| Regular expressions         | `=~"^[a-z0-9-]+$"`               |
| Literal values              | `8080`, `"info"`                 |
| Conjunction and disjunction | `uint & >0`, `int \| string`     |
| Lists and maps of the above | `[...string]`, `{[string]: string}` |
| Definitions from imports    | `schema.#ImagePullPolicy`, `odin.#Name` |

Most expressions can be written unquoted (`!type uint & >0`,
`!type =~"^[a-z]+$"`). Map types must be quoted, because an unquoted
`{...}` is a YAML flow mapping: `!type '{[string]: string}'`. All of
these were checked with `yaml.v3`, including keys with `?` and `!`
markers and head comments.

Other references, function calls, comprehensions and default markers
(`*value`) are rejected. A reference is allowed only when it is an
import alias followed by a definition. The loader parses each `!type` with CUE's
expression parser and checks the syntax tree against the allowed set
before generating code, so the rule is enforced, not documented.

- **Defaults** go in `defaults` (0005), never in `!type`. This follows
  0005's deprecation of `*value` for config defaults.
- **Named types from modules** are definitions referenced through an
  import alias: `imagePullPolicy?: !type schema.#ImagePullPolicy`, or
  combined with other constraints (`!type schema.#Port & >1024`).
- **Anything beyond the subset** (builtin validators such as
  `strings.HasSuffix`, cross-field constraints) goes in a `.cue` file
  alongside the YAML (see *Mixing YAML and CUE*).

**`!enum [a, b, c]`** — shorthand for a disjunction of literals. Writing
the CUE form in YAML needs awkward quoting (`!type '"debug" | "info"'`),
and enums are the most common constraint, so they get their own tag.

**Field markers.** A key ending in `?` is optional and a key ending in
`!` is required, as in CUE: `replicas?: !type uint`, `image!: !type
string`. An unmarked key is a regular field. The markers are valid at
the end of a plain YAML key and mean nothing to YAML itself.

**Documentation.** A YAML comment directly above a key becomes that
field's CUE doc comment, so it shows up in `odin show values`,
`odin docs` and the GUI. `yaml.v3` keeps these as head comments.

Concrete values and constraints mix freely, exactly as they do in CUE:
`port: 8080` pins a value and `image!: !type string` leaves it to the
deployer.

#### Alternatives considered for schema

- **Embedded JSON Schema** (`values: schema: {type: object, ...}`),
  converted with CUE's `encoding/jsonschema`. GUIs speak JSON Schema
  natively, but this introduces a second schema language whose semantics
  (`allOf`, `additionalProperties`) don't map one-to-one onto
  unification, and an authored JSON Schema would diverge from what
  `odin show values` reports. GUIs get JSON Schema from the schema
  export command instead.
- **A structured YAML type language** (`!type {kind: int, min: 1}`).
  Easy for a GUI to model, but it's a new language to learn and to
  maintain. The CUE subset is small enough for a GUI to model fully,
  which is the property that matters for lossless round-trips.

### Path grammar

```
path     := ( "components" "." name ( "." surface ( "." field )* )?
            | "values" ( "." field )+
            | "metadata" ( "." field )+
            | alias ( "." field )* ) call?
alias    := an identifier declared in `imports`
call     := "(" ( literal ( "," literal )* )? ")"
literal  := string | number | bool
surface  := "metadata" | "config" | "ref" | "extension" | "#" ident
name     := ident | quoted
field    := ident | quoted | "#" ident
```

- `components.<name>` with no surface is shorthand for
  `components.<name>.ref`.
- `values.components.*` is rejected. Use `components.<name>.config.*`,
  which is the merged result, so each value has exactly one path.
- `extension` gives YAML the identity helpers from 0002 and 0004. Use
  `!ref components.api.extension.resourceName("cache")` rather than
  interpolating `metadata.instance` directly: `instance` is null in
  single-instance bundles and YAML has no conditionals to guard it,
  while `resourceName` handles the null case.
- A path may end in a function call with literal arguments only. Calls
  keep references statically analyzable: the call target is a path and
  the arguments are constants. Inside `!fmt`, single-quote the YAML
  scalar so the string argument needs no escaping:
  `!fmt '{components.api.extension.resourceName("cache")}.svc'`.
- `resources`, `traits`, `implements`, `apiVersion`, `kind` and
  `_hidden` fields of a component are not referenceable. Only the public
  surface is.

### Where tags may appear

| Tag                   | `components` | `values` | `defaults` | `-f` overlays |
| --------------------- | ------------ | -------- | ---------- | ------------- |
| `!ref`, `!fmt`        | yes          | yes, limited roots | no | no          |
| `!type`, `!enum`      | yes          | yes      | no         | no            |
| Field markers `?` `!` | yes          | yes      | no         | no            |

**References in `values`** let components take values from global ones,
e.g. `values.components.myapp.hostname: !fmt "shop.{values.domain}"`.
Inside `values`, a path may only start at `values.*` (excluding
`values.components.*`, as everywhere), `metadata.*` or an import
alias. Wiring between
components, with `components.*` paths, belongs in the `components`
section. That keeps the reference graph one-directional: values can
derive from global values and bundle metadata, and components can derive
from anything, but values never depend on components.

A value bound by a reference is fixed, just as in CUE: a deployer
overlay that sets `hostname` to something else conflicts with it. To get
"defaults to the global value, but deployers can override it", the
reference would have to go in `defaults` (see *Open questions*).

Overlay files (`-f`) stay plain data, so an overlay cannot rewire a
bundle. Schema tags carry no references, so allowing them in `values`
doesn't affect static analysis. They are also
allowed in `components.<name>.config`, to narrow a template's config for
one bundle. `defaults` and overlay files are plain data.

### Compilation

1. Parse with the `yaml.v3` Node API, which keeps tags and positions.
2. Statically validate every path against the bundle schema, before
   CUE evaluation, so a typo produces "component `db` has no
   declaration `#hots`" rather than an evaluation error.
3. Generate a `cue/ast` file in the bundle's package, with one import
   declaration per `imports` entry plus `odin`. Every generated node
   carries the position of the YAML node it came from, so CUE errors
   point at the YAML.
4. Avoid CUE scoping surprises. Inside a component, `metadata` would
   resolve to the component's own metadata, and a config field named
   `values` would shadow the bundle's. So references are emitted
   through file-level aliases that cannot be shadowed:

   ```cue
   import workload "example.com/platform/workload"

   let _odinComponents = components
   let _odinValues = values
   let _odinMetadata = metadata
   let _odinImport_workload = workload
   // !ref components.api.config.port → _odinComponents.api.config.port
   // template: workload.#Deployment  → _odinImport_workload.#Deployment
   ```

   Import aliases need this as much as the bundle roots do: a config
   field named `workload` would otherwise hide the import inside that
   struct. Checked on CUE v0.18.0-rc.3: with a sibling field `lib`,
   `lib.#Deployment` fails with "undefined field", while the `let`
   alias resolves.

5. Hand the result to the normal loader path.

CUE evaluates lazily, so references need no ordering. Cycles between
YAML references are reported from the static reference graph. Cycles
inside CUE, such as a trait deriving a config field from itself (0007),
are not visible to that graph and are reported by CUE.

### Mixing YAML and CUE

A bundle package may contain `bundle.yaml` and `.cue` files together.
All of them unify into one package. GUI users stay in YAML, and anyone
who needs more power adds a `.cue` file. A GUI should show
CUE-contributed parts as read-only.

### JSON equivalence

For tools that work in JSON, each tag has a defined object form:
`{"$ref": ...}`, `{"$fmt": ...}`, `{"$type": ...}`, `{"$enum": [...]}`,
with `path` / `with` for the mapping form. The loader accepts these only in `.json` bundle
files, so the `$ref` collision in YAML values does not arise.

### Distribution

`odin push` publishes the YAML as authored. Consumers need an Odin
version that understands YAML bundles. The `v1alpha2` apiVersion check
already produces a clear error on older versions. A bundle that imports
a YAML bundle relies on its resolver compiling that YAML (see *YAML and
CUE bundles import each other*), so the same version requirement applies
to everything in a composition graph.

## Tooling surface (appendix)

The commands a GUI needs. Output formats are decided in the
implementing change, not here.

- **Schema export.** Machine-readable (JSON Schema or similar) export
  of the bundle's values schema, including constraints written with
  `!type` and `!enum`, and of a template's config schema and its public declarations, including `@odin(ref)` /
  `@odin(ext)` categories and their unfilled fields (the `with` form
  for a constructor like `#mount`). Attributes outside the `@odin`
  namespace (for example `@ui(...)`) are included as-is and not
  interpreted, so GUIs and platform modules can define their own hints.
- **Imports.** List importable packages (template modules from the
  bundle's dependencies, as `odin components` does today, and bundle
  packages), and add or remove `imports` entries. A GUI picking a
  template or trait from a package it hasn't imported yet adds the
  import for the author.
- **Compatible traits.** For a component, list every reachable trait
  (component declarations and definitions in imported packages) whose
  `target` matches the
  component's own GVK or an entry in its `implements` (0007). This is a
  lookup on concrete data, with no evaluation.
- **Compatible targets.** For any other field path, list every
  reachable reference whose type unifies with it. This one requires
  evaluation.
- **Validate.** Errors with YAML file, line, column and path.
- **Reference graph.** The static reference graph between components,
  for drawing and for checking renames.

## Open questions

- **Only annotated declarations as `#` targets?** Should `!ref` accept
  only declarations marked with `@odin(...)` (the existing public
  marker used by `odin docs`), or any non-private definition? The same
  question applies to definitions reached through an import alias. Requiring
  the annotation makes the public surface explicit and matches what
  `odin docs` already shows.
- **Exact versions in YAML.** Major versions are handled by the import
  path. Should `imports` also accept an exact version
  (`workload: {path: example.com/platform/workload@v1, version: v1.2.0}`),
  with Odin keeping `cue.mod/module.cue` in sync? A GUI would then need
  to edit only one file. Is that worth Odin owning `module.cue`, or do
  exact versions stay there?
- **Overlay compilation for local packages.** Compiling `bundle.yaml`
  into overlay files works for packages Odin loads from disk. Whether
  CUE's loader honors overlays for packages it reads from the module
  cache is unverified, which is why fetched bundles are compiled by the
  resolver instead. Prototype both paths before accepting.
- **Package name.** A YAML file has no `package` clause. Take it from
  sibling `.cue` files and default to `bundle` when there are none?
- **References in `defaults`.** The common need is likely "default to a
  global value, still overridable", which a reference in `values` can't
  express. Allowing `!ref` / `!fmt` in `defaults` with the same limited
  roots would cover it, but it interacts with 0005's loader walk: a
  default that refers to a global value which is itself filled from
  `defaults` needs the walk to resolve in dependency order. Decide
  together with 0005's open questions.
- **Literal `?` and `!` at the end of keys.** A real key ending in `?`
  or `!` would be read as a field marker. Is that rare enough to ignore,
  or does the format need an escape (for example a `!key` tag)?
- **Builtin validators.** Should `!type` allow a short list of CUE
  builtins (`strings.MinRunes`, `strings.HasPrefix`, `list.MinItems`)?
  They need imports, which the generated code can add, and they're
  common in real schemas, such as the `Mi`/`Gi` suffix check on memory
  values.
- **`!fmt` scope.** Is interpolation over any path useful, or should it
  be limited to `metadata.*` and `values.*`?

## Consequences

To be populated once the open questions are settled.

## References

- [0001 — Bundle composition](0001-bundle-composition.md)
- [0005 — Bundle defaults layering](0005-bundle-defaults-layering.md)
- [0006 — `v1alpha2` schema cohort](0006-v1alpha2-cohort.md)
- [0007 — Standard component wiring: handles, traits and extensions](0007-component-handles-and-traits.md)
