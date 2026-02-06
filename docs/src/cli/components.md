# odin components

List available component templates from bundle dependencies.

## Synopsis

```bash
odin components [location] [flags]
```

## Description

The `components` command discovers and lists all component templates available in your bundle's dependencies. It shows:

- Package path
- Definition name (e.g., `#Component`)
- Module name
- Version

This helps you discover what templates you can use in your bundle.

## Examples

List templates in current bundle:

```bash
odin components
```

List templates in specific bundle:

```bash
odin components ../my-bundle
```

Output as JSON:

```bash
odin components -f json
```

## Flags

### `-f, --format string`

Output format: `table` or `json` (default: `table`)

**Table format:**

```
PACKAGE                                              NAME         MODULE                                VERSION
platform.go-valkyrie.com/webapp/v1alpha1             #Component   platform.go-valkyrie.com/webapp       v0.1.0
platform.go-valkyrie.com/database/v1alpha1           #Component   platform.go-valkyrie.com/database     v0.1.0
```

**JSON format:**

```json
[
  {
    "package": "platform.go-valkyrie.com/webapp/v1alpha1",
    "name": "#Component",
    "module": "platform.go-valkyrie.com/webapp",
    "version": "v0.1.0"
  },
  {
    "package": "platform.go-valkyrie.com/database/v1alpha1",
    "name": "#Component",
    "module": "platform.go-valkyrie.com/database",
    "version": "v0.1.0"
  }
]
```

## Location Argument

The `[location]` argument specifies the bundle directory (default: current directory).

## How It Works

The command:

1. Loads the bundle's CUE module
2. Reads dependencies from `cue.mod/module.cue`
3. For each dependency, loads all packages
4. Finds definitions that extend `#ComponentBase`
5. Displays them in the requested format

## Bundle Auto-discovery

Like other commands, `odin components` automatically discovers the bundle root when run from a subdirectory.

## Using Discovered Templates

Once you find a template, you can:

### 1. View Documentation

```bash
odin docs <template-name>
```

### 2. Import in Bundle

Add to your `bundle.cue`:

```cue
import (
    webapp "platform.go-valkyrie.com/webapp/v1alpha1"
)

odin.#Bundle & {
    components: {
        myapp: webapp.#Component & {
            metadata: name: "myapp"
        }
    }
}
```

### 3. Configure Values

```cue
values: {
    components: myapp: {
        image: name: "myapp"
        image: tag: "v1.0.0"
    }
}
```

## Adding Dependencies

If you want to use a template not yet in your dependencies, add it to `cue.mod/module.cue`:

```cue
module: "example.com/my-bundle@v0"
language: version: "v0.11.0"

deps: {
    "go-valkyrie.com/odin/api/v1alpha1@v0": {
        v: "v0.3.0"
    }
    // Add new dependency
    "platform.go-valkyrie.com/webapp/v1alpha1@v0": {
        v: "v0.1.0"
    }
}
```

Then run:

```bash
cue mod tidy
```

## Filtering and Searching

Use standard shell tools to filter output:

```bash
# Find all webapp templates
odin components | grep webapp

# Count available templates
odin components -f json | jq 'length'

# Get specific field
odin components -f json | jq -r '.[].module' | sort -u
```

## Next Steps

- Use [odin docs](./docs.md) to view template documentation
- Learn about [Component Templates](../concepts/templates.md)
- See [Using Templates](../guides/using-templates.md) guide

## See Also

- [odin docs](./docs.md)
- [Component Templates](../concepts/templates.md)
- [Configuration](../guides/configuration.md)
