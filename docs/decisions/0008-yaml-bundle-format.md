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

- expresses everything a typical bundle file does, including template
  references and wiring between components;
- can be analyzed without evaluating CUE, so a GUI can show the
  reference graph, list valid targets and detect cycles;
- adds no semantics beyond CUE's own, and none beyond the standard
  fields of 0007;
- compiles to CUE, which remains the only evaluation engine.

## Non-goals

- **YAML component templates.** Templates remain CUE.
- **A general expression language in YAML.** No arithmetic,
  conditionals or comprehensions. Bundles that need them use CUE (see
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

`!ref`, `!def` and `!fmt`, each restricted to paths and plain data.

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
components:
  mygateway:
    template: example.com/platform/gateway#Gateway
  configmap:
    template: example.com/platform/config#ConfigMap
    config:
      data: { LOG_LEVEL: info }
  myapp:
    template: example.com/platform/workload#Deployment
    traits:
      - !ref components.configmap.#env
      - !ref
        path: components.configmap.#mount
        with:
          path: /config
      - !def example.com/platform/security#SecureDefaults
  myroute:
    template: example.com/platform/gateway#HTTPRoute
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
      imagePullPolicy?: !def platform.example.com/common/schema#ImagePullPolicy
defaults:
  components:
    myapp:
      logLevel: info
```

Top-level keys follow `#Bundle`: `apiVersion`, `kind`, `metadata`,
`components`, `values`, and `defaults` (0005).

A component accepts exactly these keys; anything else is an error:

| Key        | Maps to                                       |
| ---------- | --------------------------------------------- |
| `template` | the template definition the component unifies with |
| `metadata` | `metadata`                                    |
| `config`   | `config`                                      |
| `traits`   | `traits` (0007)                               |
| `extension` | `extension` (0007) — set extension points directly, without a trait |
| `defs`     | `#<name>` for each entry — sets arbitrary definition fields for patterns that aren't standardized |

`ref` and `implements` are set by templates, never by bundles, so they
are not component keys.

`defs` keeps the format generic without requiring quoted `"#name"`
keys. With 0007 standardizing extension points, it is expected to be
rare. An unquoted `#` at the start of a YAML token begins a comment,
which would silently drop the field.

### Template references

`template: <import path>#<Definition>`, for example
`example.com/platform/workload#WebApp`. This matches how Odin already
identifies templates (package import path plus definition name, see
`ComponentTemplate`). The loader generates one import per distinct
package, with collision-free aliases.

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
may itself contain `!ref`, `!def` and `!fmt`. Closed definitions reject
unknown `with` keys, so typos are caught.

**`!def <import path>#<Definition>`** — a definition from a module, not
from a component. It uses the same syntax as `template:` and has the
same scalar and mapping (`path` / `with`) forms. It covers traits that
don't derive from another component, such as a standard
security-defaults trait.

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

Most expressions can be written unquoted (`!type uint & >0`,
`!type =~"^[a-z]+$"`). Map types must be quoted, because an unquoted
`{...}` is a YAML flow mapping: `!type '{[string]: string}'`. All of
these were checked with `yaml.v3`, including keys with `?` and `!`
markers and head comments.

References, imports, function calls, comprehensions and default markers
(`*value`) are rejected. The loader parses each `!type` with CUE's
expression parser and checks the syntax tree against the allowed set
before generating code, so the rule is enforced, not documented.

- **Defaults** go in `defaults` (0005), never in `!type`. This follows
  0005's deprecation of `*value` for config defaults.
- **Named types from modules** use `!def`, which already exists for
  traits: `imagePullPolicy?: !def example.com/platform/schema#ImagePullPolicy`.
  Whether a definition is used as a value or a constraint makes no
  difference to CUE.
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
path     := "components" "." name ( "." surface ( "." field )* )?
          | "values" ( "." field )+
          | "metadata" ( "." field )+
surface  := "metadata" | "config" | "ref" | "#" ident
name     := ident | quoted
field    := ident | quoted | "#" ident
```

- `components.<name>` with no surface is shorthand for
  `components.<name>.ref`.
- `values.components.*` is rejected. Use `components.<name>.config.*`,
  which is the merged result, so each value has exactly one path.
- `resources`, `traits`, `extension`, `implements`, `apiVersion`,
  `kind` and `_hidden` fields of a component are not referenceable. Only the public surface is.

### Where tags may appear

| Tag                   | `components` | `values` | `defaults` | `-f` overlays |
| --------------------- | ------------ | -------- | ---------- | ------------- |
| `!ref`, `!fmt`        | yes          | yes, limited roots | no | no          |
| `!def`                | yes          | yes      | no         | no            |
| `!type`, `!enum`      | yes          | yes      | no         | no            |
| Field markers `?` `!` | yes          | yes      | no         | no            |

**References in `values`** let components take values from global ones,
e.g. `values.components.myapp.hostname: !fmt "shop.{values.domain}"`.
Inside `values`, a path may only start at `values.*` (excluding
`values.components.*`, as everywhere) or `metadata.*`. Wiring between
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
3. Generate a `cue/ast` file in the bundle's package. Every generated
   node carries the position of the YAML node it came from, so CUE
   errors point at the YAML.
4. Avoid CUE scoping surprises. Inside a component, `metadata` would
   resolve to the component's own metadata, and a config field named
   `values` would shadow the bundle's. So references are emitted
   through file-level aliases that cannot be shadowed:

   ```cue
   let _odinComponents = components
   let _odinValues = values
   let _odinMetadata = metadata
   // !ref components.api.config.port → _odinComponents.api.config.port
   ```

5. Hand the result to the normal loader path.

CUE evaluates lazily, so references need no ordering. Real cycles are
reported from the static reference graph.

### Mixing YAML and CUE

A bundle package may contain `bundle.yaml` and `.cue` files together.
All of them unify into one package. GUI users stay in YAML, and anyone
who needs more power adds a `.cue` file. A GUI should show
CUE-contributed parts as read-only.

### JSON equivalence

For tools that work in JSON, each tag has a defined object form:
`{"$ref": ...}`, `{"$def": ...}`, `{"$fmt": ...}`, with `path` / `with`
for the mapping form. The loader accepts these only in `.json` bundle
files, so the `$ref` collision in YAML values does not arise.

### Distribution

`odin push` publishes the YAML as authored. Consumers need an Odin
version that understands YAML bundles. The `v1alpha2` apiVersion check
already produces a clear error on older versions.

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
- **Compatible traits.** For a component, list every reachable trait
  (component declarations and known `!def`s) whose `target` matches the
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
  marker used by `odin docs`), or any non-private definition? Requiring
  the annotation makes the public surface explicit and matches what
  `odin docs` already shows.
- **Module versions in YAML.** Should `template:` and `!def` accept a
  version (`example.com/platform@v1.2.0/workload#WebApp`), with Odin
  managing `cue.mod/module.cue`? A GUI would then need to edit only
  one file.
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
