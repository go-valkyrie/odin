# Bundle Structure

This guide covers best practices for organizing Odin bundles.

## Basic Structure

A typical bundle follows this layout:

```
my-bundle/
├── bundle.cue              # Main bundle definition
├── components/             # Component definitions
│   ├── webapp.cue
│   ├── database.cue
│   └── cache.cue
├── lib/                    # Shared definitions and helpers
│   ├── labels.cue
│   └── common.cue
├── environments/           # Environment-specific values
│   ├── dev.cue
│   ├── staging.cue
│   └── prod.cue
├── values.cue              # Default values
├── odin.toml               # Bundle-specific config (optional)
├── cue.mod/
│   └── module.cue          # CUE module definition
├── manifests/              # Rendered manifests (gitops)
│   ├── dev.yaml
│   ├── staging.yaml
│   └── prod.yaml
└── README.md
```

## File Organization

### Single File vs. Multiple Files

**Small bundles** - Single file is fine:

```cue
// bundle.cue
package main

import odin "go-valkyrie.com/odin/api/v1alpha1"

odin.#Bundle & {
    components: {
        webapp: {...}
        database: {...}
    }
    values: {...}
}
```

**Large bundles** - Split into multiple files:

```cue
// bundle.cue
package main

import (
    odin "go-valkyrie.com/odin/api/v1alpha1"
    "example.com/my-bundle/components"
)

odin.#Bundle & {
    metadata: name: "my-bundle"
    components
}
```

```cue
// components/webapp.cue
package components

import odin "go-valkyrie.com/odin/api/v1alpha1"

webapp: odin.#Component & {
    // Component definition
}
```

### Package Names

All files in a bundle typically use the same package:

```cue
// bundle.cue
package main

// components/webapp.cue
package main

// values.cue
package main
```

CUE automatically merges files with the same package name.

## Component Organization

### Inline Components

For simple bundles:

```cue
odin.#Bundle & {
    components: {
        webapp: odin.#Component & {
            metadata: name: "webapp"
            config: {...}
            resources: {...}
        }
    }
}
```

### Separate Component Files

For better organization:

```cue
// components/webapp.cue
package main

import odin "go-valkyrie.com/odin/api/v1alpha1"

components: {
    webapp: odin.#Component & {
        metadata: name: "webapp"
        config: {...}
        resources: {...}
    }
}
```

```cue
// bundle.cue
package main

import odin "go-valkyrie.com/odin/api/v1alpha1"

odin.#Bundle & {
    metadata: name: "my-bundle"
    // components defined in components/*.cue are automatically included
}
```

### Component Packages

For complex components:

```
components/
├── webapp/
│   ├── component.cue       # Main component
│   ├── deployment.cue      # Deployment resource
│   ├── service.cue         # Service resource
│   └── ingress.cue         # Ingress resource (optional)
└── database/
    ├── component.cue
    ├── statefulset.cue
    └── configmap.cue
```

## Values Organization

### Default Values

Put common defaults in `values.cue`:

```cue
// values.cue
package main

values: {
    domain: "example.com"
    namespace: "default"
    
    components: {
        webapp: {
            replicas: 1
            image: {
                repository: "ghcr.io/myorg/webapp"
                tag: "latest"
            }
        }
    }
}
```

### Environment Values

Create environment-specific files:

```cue
// environments/dev.cue
package main

values: {
    domain: "dev.example.com"
    
    components: {
        webapp: {
            replicas: 1
            resources: {
                requests: {
                    memory: "64Mi"
                    cpu: "50m"
                }
            }
        }
    }
}
```

```cue
// environments/prod.cue
package main

values: {
    domain: "example.com"
    
    components: {
        webapp: {
            replicas: 5
            image: tag: "v1.0.0"  // Pin version in prod
            resources: {
                requests: {
                    memory: "512Mi"
                    cpu: "500m"
                }
            }
        }
    }
}
```

## Shared Code

Create reusable definitions in `lib/`:

```cue
// lib/labels.cue
package lib

#CommonLabels: {
    app: string
    environment: string
    version: string
    team: "platform"
}
```

```cue
// components/webapp.cue
package main

import "example.com/my-bundle/lib"

components: {
    webapp: odin.#Component & {
        #labels: lib.#CommonLabels & {
            app: "webapp"
            environment: values.environment
            version: config.image.tag
        }
        
        resources: {
            deployment: metadata: labels: #labels
            service: metadata: labels: #labels
        }
    }
}
```

## GitOps Structure

For GitOps workflows, render manifests to a `manifests/` directory:

```
my-bundle/
├── ...
├── manifests/
│   ├── dev/
│   │   ├── kustomization.yaml
│   │   └── all.yaml
│   ├── staging/
│   │   ├── kustomization.yaml
│   │   └── all.yaml
│   └── prod/
│       ├── kustomization.yaml
│       └── all.yaml
└── scripts/
    └── render.sh
```

**scripts/render.sh:**

```bash
#!/bin/bash
set -euo pipefail

for env in dev staging prod; do
    echo "Rendering $env..."
    odin template . -f "environments/${env}.cue" > "manifests/${env}/all.yaml"
done

echo "Done!"
```

## Module Configuration

### Basic module.cue

```cue
module: "example.com/my-bundle@v0"
language: version: "v0.11.0"

deps: {
    "go-valkyrie.com/odin/api/v1alpha1@v0": {
        v: "v0.3.0"
    }
}
```

### With External Templates

```cue
module: "example.com/my-bundle@v0"
language: version: "v0.11.0"

deps: {
    "go-valkyrie.com/odin/api/v1alpha1@v0": {
        v: "v0.3.0"
    }
    "platform.go-valkyrie.com/webapp/v1alpha1@v0": {
        v: "v0.1.0"
    }
    "platform.go-valkyrie.com/database/v1alpha1@v0": {
        v: "v0.1.0"
    }
}
```

## Documentation

Include a README with:

- Bundle purpose
- How to render manifests
- Environment descriptions
- Component inventory
- Contact information

Example:

```markdown
# My App Bundle

Odin bundle for deploying My App to Kubernetes.

## Components

- **webapp** - Main web application (Go + React)
- **database** - PostgreSQL database
- **cache** - Redis cache

## Environments

- **dev** - Development environment
- **staging** - Staging environment
- **prod** - Production environment

## Usage

Render manifests:

```bash
# Development
odin template . -f environments/dev.cue

# Production
odin template . -f environments/prod.cue
```

## Configuration

See `values.cue` for default values and component configuration options.
```

## Multi-Application Bundles

You can have multiple applications in one bundle:

```
my-platform/
├── bundle.cue
├── apps/
│   ├── frontend/
│   │   ├── component.cue
│   │   └── values.cue
│   ├── backend/
│   │   ├── component.cue
│   │   └── values.cue
│   └── worker/
│       ├── component.cue
│       └── values.cue
├── shared/
│   ├── ingress.cue
│   └── monitoring.cue
├── environments/
│   ├── dev.cue
│   ├── staging.cue
│   └── prod.cue
└── cue.mod/
```

## Anti-Patterns

### Don't: Nested cue.mod

```
❌ my-bundle/
   ├── cue.mod/
   └── components/
       └── webapp/
           └── cue.mod/    # Don't do this
```

Each bundle should have one `cue.mod/` at the root.

### Don't: Mix Package Names

```cue
❌ // bundle.cue
   package main

❌ // components/webapp.cue
   package webapp    # Different package!
```

Use consistent package names (usually `main`).

### Don't: Duplicate Values

```cue
❌ // Anti-pattern: duplicating values
   values: {
       components: {
           webapp: replicas: 3
           api: replicas: 3
           worker: replicas: 3
       }
   }

✅ // Better: shared values
   values: {
       defaultReplicas: 3
       components: {
           webapp: replicas: values.defaultReplicas
           api: replicas: values.defaultReplicas
           worker: replicas: values.defaultReplicas
       }
   }
```

## Next Steps

- Learn [Writing Components](./writing-components.md)
- See [Working with Values](./values.md)
- Explore [Best Practices](../reference/best-practices.md)
