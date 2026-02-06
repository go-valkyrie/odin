# odin config

Utilities for working with Odin configuration.

## Synopsis

```bash
odin config <subcommand> [flags]
```

## Description

The `config` command provides utilities for managing Odin's configuration file at `~/.config/odin/config.cue`.

## Subcommands

### eval

Display the evaluated configuration:

```bash
odin config eval
```

Shows the current configuration with all values resolved.

## Configuration File

Odin looks for configuration at `~/.config/odin/config.cue`:

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

## Configuration Schema

The configuration file supports:

### Registries

Configure CUE registry mappings:

```cue
registries: [
    {
        prefix: "platform.example.com"
        registry: "ghcr.io/myorg/cue"
    },
    {
        prefix: "internal.company.com"
        registry: "registry.company.com/cue"
    },
]
```

When Odin loads a module with a matching prefix, it will use the specified OCI registry.

## Configuration Precedence

Odin uses a three-tier configuration system:

1. **User config** - `~/.config/odin/config.cue`
2. **Bundle config** - `odin.toml` in bundle root
3. **Default registries** - Built-in defaults

Later tiers override earlier ones.

## Bundle Configuration (odin.toml)

You can also configure registries per-bundle with `odin.toml`:

```toml
[[registries]]
prefix = "platform.example.com"
registry = "ghcr.io/myorg/cue"
```

This is useful for project-specific registry configuration.

## Environment Variables

Override registry configuration with:

```bash
export CUE_REGISTRY=/path/to/registryconfig.cue
odin template .
```

## Examples

View current configuration:

```bash
odin config eval
```

Example output:

```cue
{
    registries: [
        {
            prefix: "go-valkyrie.com"
            registry: "ghcr.io/go-valkyrie/cue"
        },
        {
            prefix: "platform.example.com"
            registry: "ghcr.io/myorg/cue"
        },
    ]
}
```

## Next Steps

- Learn about [Configuration & Registries](../guides/configuration.md)
- See [Publishing Components](../guides/publishing.md) for registry setup

## See Also

- [Configuration Guide](../guides/configuration.md)
- [odin cache](./cache.md)
