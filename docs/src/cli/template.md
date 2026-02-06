# odin template

Render Kubernetes manifests from an Odin bundle.

## Synopsis

```bash
odin template [location] [flags]
```

## Description

The `template` command renders Kubernetes manifests from an Odin bundle. It:

1. Loads the bundle from the specified location
2. Merges any provided values files
3. Evaluates all components
4. Outputs YAML manifests to stdout

The output can be piped to `kubectl apply`, saved to files, or used with ArgoCD.

## Examples

Render bundle in current directory:

```bash
odin template .
```

Render with custom values:

```bash
odin template . -f environments/prod.cue
```

Multiple values files (merged in order):

```bash
odin template . -f base.cue -f prod.cue -f override.cue
```

Save to file:

```bash
odin template . > manifests.yaml
```

Apply directly to cluster:

```bash
odin template . | kubectl apply -f -
```

Render from subdirectory:

```bash
cd components/
odin template .  # Automatically finds bundle root
```

## Flags

### `-f, --values stringArray`

Specify one or more values files to merge with the bundle.

Values files must be CUE files in the same package as the bundle (typically `package main`). Later files override earlier ones.

```bash
# Single values file
odin template . -f prod-values.cue

# Multiple values files (merged left to right)
odin template . -f base.cue -f env/prod.cue -f regional/us-west.cue
```

## Location Argument

The `[location]` argument specifies where to find the bundle. It can be:

### Local Path

```bash
# Current directory
odin template .

# Specific directory
odin template ../my-bundle

# Absolute path
odin template /path/to/bundle
```

### Future: Remote Sources

Remote sources (OCI, git) are planned for future releases:

```bash
# OCI registry (future)
odin template oci://ghcr.io/myorg/bundle:v1.0.0

# Git repository (future)
odin template git://github.com/myorg/bundle.git?ref=main
```

## Bundle Auto-discovery

When you run `odin template` from within a bundle, Odin automatically finds the bundle root by walking up the directory tree to find `cue.mod/`.

This means you can run the command from any subdirectory:

```
my-bundle/
├── bundle.cue
├── components/
│   └── webapp.cue      ← You are here
└── cue.mod/

$ odin template .
# Works! Finds bundle root automatically
```

## Output Format

The command outputs YAML documents separated by `---`:

```yaml
---
# Source: component-name/resource-name
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  ...
---
# Source: component-name/another-resource
apiVersion: v1
kind: Service
metadata:
  name: myapp
spec:
  ...
```

The `# Source:` comment indicates which component and resource generated each manifest.

## Resource Ordering

Resources are output in deterministic order:

1. Components are sorted alphabetically by name
2. Within each component, resources are sorted alphabetically by key

This ensures consistent output across runs.

## Values Merging

When multiple values files are provided with `-f`, they are merged in order from left to right:

```bash
odin template . -f a.cue -f b.cue -f c.cue
```

Merge order: `bundle.cue` → `a.cue` → `b.cue` → `c.cue`

Later files can override earlier values:

**base.cue:**
```cue
package main
values: components: webapp: replicas: 1
```

**prod.cue:**
```cue
package main
values: components: webapp: replicas: 5
```

Result: `replicas: 5`

## Error Handling

The command will exit with an error if:

- The bundle location doesn't exist
- Bundle has CUE syntax errors
- Values fail to unify with component schemas
- Required configuration fields are missing
- Type constraints are violated

Example error:

```
Error: value "invalid" not an instance of "dev"|"staging"|"prod"
    bundle.cue:15:9
```

## Performance

For large bundles:

- Use `--debug` to see what's being loaded
- Check cache status with `odin cache clean --dry-run`
- Ensure `cue.mod/` doesn't have unnecessary files

## Use Cases

### GitOps Workflow

```bash
# Render manifests for each environment
odin template . -f env/dev.cue > manifests/dev.yaml
odin template . -f env/staging.cue > manifests/staging.yaml
odin template . -f env/prod.cue > manifests/prod.yaml

# Commit to git
git add manifests/
git commit -m "Update manifests"
```

### CI/CD Pipeline

```bash
#!/bin/bash
# render-manifests.sh

set -euo pipefail

ENV=${1:?environment required}

odin template . -f "environments/${ENV}.cue" | kubectl apply --dry-run=client -f -
```

### Local Development

```bash
# Render and apply in one command
odin template . -f dev.cue | kubectl apply -f -

# Watch for changes (with entr or similar)
find . -name '*.cue' | entr -r sh -c 'odin template . | kubectl apply -f -'
```

## Next Steps

- Learn about [Values](../concepts/values.md) for configuration management
- See [GitOps Workflows](../integration/gitops.md) for deployment patterns
- Explore [Best Practices](../reference/best-practices.md) for bundle organization

## See Also

- [Bundles](../concepts/bundles.md)
- [Values](../concepts/values.md)
- [odin init](./init.md)
