# Configuration File Format

This reference documents Odin's configuration file formats.

## User Configuration

**Location:** `~/.config/odin/config.cue`

```cue
package config

import "go-valkyrie.com/odin/internal/config"

config.#Config & {
    registries: [
        {
            prefix: "platform.example.com"
            registry: "ghcr.io/myorg/cue"
        },
    ]
}
```

### Schema

**registries** (array) - Registry mappings

Each registry entry has:

- `prefix` (string, required) - Module path prefix to match
- `registry` (string, required) - OCI registry URL

Example:

```cue
registries: [
    {
        // Modules starting with "platform.example.com"
        prefix: "platform.example.com"
        // Are fetched from this registry
        registry: "ghcr.io/myorg/cue"
    },
    {
        prefix: "internal.company.com"
        registry: "registry.company.com/cue"
    },
]
```

## Bundle Configuration

**Location:** `odin.toml` in bundle root

```toml
[[registries]]
prefix = "platform.example.com"
registry = "ghcr.io/myorg/cue"

[[registries]]
prefix = "internal.company.com"
registry = "registry.company.com/cue"
```

### TOML Format

**[[registries]]** - Array of registry mappings

Each registry table has:

- `prefix` (string, required) - Module path prefix
- `registry` (string, required) - OCI registry URL

## CUE Module Configuration

**Location:** `cue.mod/module.cue`

```cue
module: "example.com/my-bundle@v0"
language: version: "v0.11.0"

deps: {
    "go-valkyrie.com/odin/api/v1alpha1@v0": {
        v: "v0.3.0"
    }
    "platform.example.com/webapp/v1alpha1@v0": {
        v: "v0.1.0"
    }
}
```

### Schema

**module** (string, required) - Module path with major version

Format: `domain/path@vN`

Examples:
- `example.com/my-bundle@v0`
- `github.com/myorg/templates@v1`

**language.version** (string, required) - CUE language version

Use the version of CUE you're targeting.

**deps** (map, optional) - Module dependencies

Keys are module paths with major version, values contain:

- `v` (string, required) - Specific version to use

Examples:

```cue
// Pin specific version
deps: {
    "platform.example.com/webapp/v1alpha1@v0": {
        v: "v0.2.5"
    }
}

// Use latest compatible version
deps: {
    "platform.example.com/webapp/v1alpha1@v0": {
        v: ">=v0.2.0"
    }
}
```

## Environment Variables

### CUE_REGISTRY

Override registry configuration file:

```bash
export CUE_REGISTRY=/path/to/registryconfig.cue
```

**registryconfig.cue:**

```cue
package cueregistry

registries: [
    {
        prefix: "platform.example.com"
        registry: "ghcr.io/myorg/cue"
    },
]
```

### XDG_CACHE_HOME

Override cache directory:

```bash
export XDG_CACHE_HOME=/custom/cache
```

Odin will use `$XDG_CACHE_HOME/odin/` for cache.

### HOME

Used to find config directory:

```bash
# Config at: $HOME/.config/odin/config.cue
export HOME=/custom/home
```

## Configuration Precedence

When Odin looks for registry configuration, it checks (in order):

1. `CUE_REGISTRY` environment variable
2. Bundle-local `odin.toml`
3. User config `~/.config/odin/config.cue`
4. Default registries (built-in)

Later sources override earlier ones.

## Examples

### Basic User Config

```cue
package config

import "go-valkyrie.com/odin/internal/config"

config.#Config & {
    registries: [
        {
            prefix: "platform.go-valkyrie.com"
            registry: "ghcr.io/go-valkyrie/cue"
        },
    ]
}
```

### Multi-Registry Config

```cue
package config

import "go-valkyrie.com/odin/internal/config"

config.#Config & {
    registries: [
        // Public templates
        {
            prefix: "platform.go-valkyrie.com"
            registry: "ghcr.io/go-valkyrie/cue"
        },
        // Internal templates
        {
            prefix: "internal.company.com"
            registry: "registry.company.com/cue"
        },
        // Team-specific templates
        {
            prefix: "team.internal.company.com"
            registry: "registry.company.com/team-cue"
        },
    ]
}
```

### Bundle with pinned dependencies

```cue
// cue.mod/module.cue
module: "example.com/my-app@v0"
language: version: "v0.11.0"

deps: {
    // Odin API
    "go-valkyrie.com/odin/api/v1alpha1@v0": {
        v: "v0.3.0"
    }
    
    // Application templates (pinned)
    "platform.example.com/webapp/v1alpha1@v0": {
        v: "v0.2.5"
    }
    "platform.example.com/database/v1alpha1@v0": {
        v: "v0.1.0"
    }
}
```

## Validation

View your effective configuration:

```bash
odin config eval
```

This shows the merged configuration from all sources.

## Next Steps

- See [Configuration Guide](../guides/configuration.md) for setup instructions
- Learn [Best Practices](./best-practices.md)
- Explore [API Schema](./api-schema.md)
