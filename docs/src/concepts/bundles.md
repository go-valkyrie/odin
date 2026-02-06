# Bundles

A **bundle** is the top-level abstraction in Odin. It represents a complete deployable unit containing one or more components and their configuration.

## What is a Bundle?

A bundle is:

- A CUE module that unifies with the `odin.#Bundle` schema
- A collection of components that work together
- A set of configurable values that customize component behavior
- The unit of deployment in your GitOps workflow

## Bundle Structure

A minimal bundle looks like this:

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
        // Components go here
    }
    values: {
        // Configuration values go here
    }
}
```

## Bundle Schema

The `#Bundle` schema has three main sections:

### Metadata

```cue
metadata: {
    name: string  // Required: bundle name
}
```

The metadata section identifies the bundle. Currently only the `name` field is required.

### Components

```cue
components: [Name=string]: #ComponentBase
```

The components section is a map where:

- Keys are component names (must be valid CUE identifiers)
- Values must unify with `#ComponentBase` (typically `#Component`)

Example:

```cue
components: {
    webapp: odin.#Component & {
        metadata: name: "webapp"
        // ... component definition
    }
    
    database: odin.#Component & {
        metadata: name: "database"
        // ... component definition
    }
}
```

### Values

```cue
values: {
    components: [string]: {...}
    ...  // Additional fields allowed
}
```

The values section provides configuration that flows into components. It has a special `components` field that automatically binds to component configs:

```cue
values: {
    components: {
        webapp: {
            image: "nginx:latest"
            replicas: 3
        }
    }
    
    // You can also define global values
    domain: "example.com"
    environment: "production"
}
```

## Bundle Auto-discovery

When you run Odin commands, you don't always need to specify the bundle location. Odin will automatically discover the bundle root by walking up from your current directory to find `cue.mod/`.

This means you can run commands from anywhere within your bundle:

```bash
# All of these work from anywhere in the bundle:
odin template .
odin components
odin docs webapp
```

## Bundle Directory Layout

A typical bundle follows this structure:

```
my-bundle/
├── bundle.cue              # Main bundle definition
├── components/             # Component definitions (optional)
│   ├── webapp.cue
│   └── database.cue
├── values.cue              # Default values (optional)
├── environments/           # Environment-specific values (optional)
│   ├── dev.cue
│   ├── staging.cue
│   └── prod.cue
└── cue.mod/
    ├── module.cue          # CUE module definition
    └── gen/                # Generated code (if using external templates)
```

## Package Organization

All files in a bundle typically use the same package name (usually `package main`). CUE will automatically merge all files with the same package name:

```cue
// bundle.cue
package main

odin.#Bundle & {
    metadata: name: "my-app"
    components: {...}
}
```

```cue
// values.cue
package main

values: {
    domain: "example.com"
}
```

Both files contribute to the same unified structure.

## Importing External Templates

Bundles can import component templates from other CUE modules:

```cue
package main

import (
    odin "go-valkyrie.com/odin/api/v1alpha1"
    webapp "platform.go-valkyrie.com/webapp/v1alpha1"
)

odin.#Bundle & {
    metadata: name: "my-app"
    
    components: {
        myapp: webapp.#Component & {
            metadata: name: "myapp"
        }
    }
    
    values: {
        components: myapp: {
            image: name: "my-app"
            image: tag: "v1.0.0"
        }
    }
}
```

See [Component Templates](./templates.md) for more details.

## Bundle as a Git Repository

A bundle is often tracked in a Git repository. This enables:

- Version control of infrastructure configuration
- GitOps workflows with ArgoCD or Flux
- Team collaboration with pull requests
- Audit trail of all changes

Example Git structure:

```
my-app-bundle/
├── .git/
├── bundle.cue
├── components/
├── environments/
│   ├── dev.yaml           # Rendered manifests
│   ├── staging.yaml
│   └── prod.yaml
├── README.md
└── cue.mod/
```

## Multiple Bundles in a Repository

You can have multiple bundles in one repository:

```
infrastructure/
├── app1/
│   ├── bundle.cue
│   └── cue.mod/
├── app2/
│   ├── bundle.cue
│   └── cue.mod/
└── shared/
    └── lib/
        └── common.cue
```

Each bundle is independent with its own `cue.mod/` directory.

## Next Steps

- Learn about [Components](./components.md) - the building blocks of bundles
- Understand [Values](./values.md) - how configuration flows through bundles
- Explore [Bundle Structure](../guides/bundle-structure.md) - best practices for organizing bundles
