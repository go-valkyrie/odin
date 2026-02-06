# odin docs

Display documentation for component templates or packages.

## Synopsis

```bash
odin docs <reference> [flags]
```

## Description

The `docs` command displays documentation for component templates. It shows:

- Configuration schema
- Required and optional fields
- Default values
- Field descriptions
- Nested structures

This helps you understand how to configure templates in your bundle.

## Examples

Show docs for a template by name:

```bash
odin docs webapp
```

Show docs using package.Definition syntax:

```bash
odin docs workload.Deployment
```

Show docs with full package path:

```bash
odin docs platform.go-valkyrie.com/webapp/v1alpha1
```

Show all templates in a package:

```bash
odin docs platform.vituity.com/common/workload
```

Output as markdown:

```bash
odin docs webapp -f markdown > webapp.md
```

Generate mdbook documentation:

```bash
odin docs mypackage -f mdbook -o docs/
```

Expand referenced definitions inline:

```bash
odin docs webapp --expand
```

## Reference Resolution

The `<reference>` argument supports multiple formats (all case-insensitive):

### 1. Definition Name

Match by definition name (without `#`):

```bash
odin docs component
odin docs webapp
```

Matches `#Component` or `#WebApp` if unique.

### 2. Package Name

Match by last path segment (without version):

```bash
odin docs gateway  # Matches gateway.#Component if package has one definition
```

### 3. Dot Syntax

Disambiguate with `package.Definition`:

```bash
odin docs workload.Deployment
odin docs gateway.Component
```

### 4. Package Path

Show all templates in a package:

```bash
odin docs platform.vituity.com/common
odin docs platform.vituity.com/common/workload
```

### 5. Fully Qualified

Use complete package path with definition:

```bash
odin docs platform.go-valkyrie.com/webapp/v1alpha1:#Component
```

## Flags

### `-b, --bundle string`

Bundle location (default: current directory)

```bash
odin docs webapp --bundle ../my-bundle
```

### `-f, --format string`

Output format (default: `text`)

**Available formats:**

- `text` - Colored terminal output (default)
- `markdown` or `md` - Single markdown document
- `markdown-multi` or `mdm` - One file per template (requires `-o`)
- `mdbook` or `mdb` - mdBook format with SUMMARY.md (requires `-o`)

### `-o, --output string`

Output file or directory path

Required for `markdown-multi` and `mdbook` formats.

**Single file (markdown):**

```bash
odin docs webapp -f markdown -o webapp.md
```

**Directory (markdown-multi):**

```bash
odin docs mypackage -f markdown-multi -o docs/
```

Creates:
```
docs/
├── Component.md
├── Deployment.md
└── Service.md
```

**Directory (mdbook):**

```bash
odin docs mypackage -f mdbook -o docs/
```

Creates:
```
docs/
├── SUMMARY.md
├── Component.md
├── Deployment.md
└── Service.md
```

### `--no-summary`

Disable SUMMARY.md generation in mdbook format

```bash
odin docs mypackage -f mdbook -o docs/ --no-summary
```

### `--expand`

Recursively expand referenced definitions inline

When a field references another definition, `--expand` includes that definition's documentation inline:

```bash
odin docs webapp --expand
```

**Without --expand:**
```
config:
  traits: []#Trait
```

**With --expand:**
```
config:
  traits: []#Trait
    enabled: bool
    priority: int
```

## Output Examples

### Text Format

```
#Component

Configuration:
  image:
    repository: string
      Container image repository
    tag: string (default: "latest")
      Image tag
  replicas: int (default: 1)
    Number of pod replicas
  resources:
    requests:
      memory: string (default: "128Mi")
      cpu: string (default: "100m")
```

### Markdown Format

````markdown
# #Component

## Configuration

### image

Container image configuration

- **repository** (string, required) - Container image repository
- **tag** (string, default: `"latest"`) - Image tag

### replicas

Number of pod replicas

- Type: int
- Default: 1
````

## Bundle Auto-discovery

The command automatically discovers the bundle root when run from a subdirectory:

```bash
cd components/
odin docs webapp  # Finds bundle root automatically
```

## Use Cases

### Quick Reference

```bash
# Check what fields a template accepts
odin docs webapp

# See all available templates in a package
odin docs platform.go-valkyrie.com/common
```

### Generate Documentation Site

```bash
# Generate mdbook documentation for all templates
odin docs myorg.com/templates -f mdbook -o docs/

# Build with mdbook
cd docs && mdbook build
```

### IDE Integration

```bash
# Generate markdown for IDE tooltip
odin docs webapp -f markdown
```

### CI/CD Documentation

```bash
#!/bin/bash
# docs-gen.sh

# Generate docs for each package
for pkg in $(odin components -f json | jq -r '.[].package' | sort -u); do
  odin docs "$pkg" -f markdown -o "docs/${pkg//\//-}.md"
done
```

## Next Steps

- Learn about [Component Templates](../concepts/templates.md)
- See [Using Templates](../guides/using-templates.md) guide
- Explore [odin components](./components.md) to discover templates

## See Also

- [odin components](./components.md)
- [Component Templates](../concepts/templates.md)
- [Publishing Components](../guides/publishing.md)
