# Compatibility Levels

Bundles may declare a `compat` level in `odin.toml` to opt into new behavior introduced
by breaking changes:

```toml
compat = 1
```

New bundles created with `odin init` always default to the latest level. Bundles without
a `compat` field are treated as `compat = 0` — this will become a warning, and
eventually a hard error, in future releases.

Removing support for a compat level is a breaking change and will be accompanied by a
major version bump.

---

## Level 0 _(legacy)_

Original behavior. No `compat` field required.

## Level 1

- **`--namespace`**: passing `--namespace` to a bundle that doesn't reference namespace
  no longer fails CUE evaluation.

### Migrating to Level 1

Update any `@tag(namespace)` usage to the `TagVars` syntax:

```cue
// before
foo: @tag(namespace)

// after
foo: @tag(namespace, var=namespace)
```

---

For the `go-valkyrie.com/odin/api` module's schema-evolution rules
(distinct from bundle compat levels), see [`api/COMPAT.md`](../api/COMPAT.md).
