# odin init

Initialize a new Odin bundle.

## Synopsis

```bash
odin init [path] [flags]
```

## Description

The `init` command creates a new Odin bundle with the necessary directory structure and files. It sets up:

- A CUE module (`cue.mod/module.cue`)
- A basic bundle definition (`bundle.cue`)
- Optional default values file (`values.cue`)

By default, the module name is inferred from the git remote URL of the repository. You can override this with the `-m` flag.

## Examples

Initialize in current directory:

```bash
odin init
```

Initialize in a new directory:

```bash
odin init my-app
```

Specify module name explicitly:

```bash
odin init --module example.com/my-app
```

Non-interactive mode:

```bash
odin init --prompt=false --module example.com/my-app
```

## Flags

### `-m, --module string`

Specify the name of the generated CUE module.

By default, Odin infers the module name from the git remote URL. For example, if your repository is `github.com/myorg/myapp`, the module will be `github.com/myorg/myapp@v0`.

### `-p, --prompt`

Use interactive prompts to configure values (default: `true`).

When enabled, Odin will prompt you for:
- Module name (with inferred default)
- Whether to create a values file
- Other initial configuration

Set to `false` for non-interactive usage:

```bash
odin init --prompt=false -m example.com/my-app
```

## Created Structure

After running `odin init`, you'll have:

```
.
├── bundle.cue
├── cue.mod/
│   └── module.cue
└── values.cue (optional)
```

### bundle.cue

A basic bundle definition:

```cue
package main

import (
    odin "go-valkyrie.com/odin/api/v1alpha1"
)

odin.#Bundle & {
    metadata: {
        name: "my-app"
    }
    components: {
        // Your components go here
    }
    values: {
        // Your values go here
    }
}
```

### cue.mod/module.cue

The CUE module definition:

```cue
module: "example.com/my-app@v0"
language: version: "v0.11.0"

deps: {
    "go-valkyrie.com/odin/api/v1alpha1@v0": {
        v: "v0.3.0"
    }
}
```

## Module Name Inference

Odin tries to infer the module name from git:

1. Runs `git remote get-url origin`
2. Parses the repository URL
3. Constructs a module path

Examples:

| Git Remote | Inferred Module |
|------------|-----------------|
| `github.com/myorg/myapp` | `github.com/myorg/myapp@v0` |
| `git@github.com:myorg/myapp.git` | `github.com/myorg/myapp@v0` |
| `https://gitlab.com/team/project.git` | `gitlab.com/team/project@v0` |

If git is not available or no remote is configured, you must provide `-m`.

## Working Directory

The `[path]` argument specifies where to initialize the bundle:

```bash
# Initialize in current directory
odin init

# Initialize in specific directory
odin init my-app

# Initialize in nested path
odin init projects/apps/webapp
```

The directory will be created if it doesn't exist.

## Next Steps

After initializing a bundle:

1. Edit `bundle.cue` to add components
2. Add dependencies to `cue.mod/module.cue` if using external templates
3. Render manifests with `odin template .`

See the [Quick Start](../getting-started/quick-start.md) guide for a complete example.

## See Also

- [Bundles](../concepts/bundles.md) - Learn about bundle structure
- [odin template](./template.md) - Render your bundle
- [Bundle Structure](../guides/bundle-structure.md) - Best practices
