# Components

A **component** is a deployable unit within a bundle. Each component represents a logical piece of your application (like a web server, database, or cache) and contains both configuration and the Kubernetes resources it generates.

## What is a Component?

A component:

- Has a name and metadata
- Defines a configuration schema (`config`)
- Generates one or more Kubernetes resources
- Receives values from the bundle's `values` section

## Component Schema

Components must unify with `odin.#Component` (or any definition that extends `#ComponentBase`):

```cue
import odin "go-valkyrie.com/odin/api/v1alpha1"

mycomponent: odin.#Component & {
    metadata: {
        name: string  // Required
        labels?: [string]: string
        annotations?: [string]: string
    }
    
    config: {
        // Your configuration schema
    }
    
    resources: {
        [string]: {
            // Kubernetes resources
        }
    }
}
```

## The Three Parts of a Component

### 1. Metadata

The metadata section identifies and tags the component:

```cue
metadata: {
    name: "webapp"
    labels: {
        tier: "frontend"
        team: "platform"
    }
    annotations: {
        "example.com/owner": "platform-team"
    }
}
```

Only `name` is required. Labels and annotations are optional but useful for organization.

### 2. Config

The config section defines what values the component accepts:

```cue
config: {
    // Required field
    image: string
    
    // Optional with default
    replicas: int | *1
    
    // Nested structure
    resources: {
        requests: {
            memory: string | *"128Mi"
            cpu: string | *"100m"
        }
        limits: {
            memory: string | *"256Mi"
            cpu: string | *"200m"
        }
    }
    
    // Map of environment variables
    env: [string]: string
}
```

CUE's type system validates configuration:

- Types are checked at build time
- Default values simplify configuration
- Constraints ensure valid configurations

### 3. Resources

The resources section contains Kubernetes manifests:

```cue
resources: {
    // Resource name is the key
    deployment: {
        apiVersion: "apps/v1"
        kind: "Deployment"
        metadata: {
            name: metadata.name
            labels: metadata.labels
        }
        spec: {
            replicas: config.replicas
            selector: matchLabels: app: metadata.name
            template: {
                metadata: labels: app: metadata.name
                spec: containers: [{
                    name: metadata.name
                    image: config.image
                    env: [
                        for k, v in config.env {
                            name: k
                            value: v
                        }
                    ]
                    resources: config.resources
                }]
            }
        }
    }
    
    service: {
        apiVersion: "v1"
        kind: "Service"
        metadata: {
            name: metadata.name
            labels: metadata.labels
        }
        spec: {
            selector: app: metadata.name
            ports: [{
                port: 80
                targetPort: 8080
            }]
        }
    }
}
```

Each key in `resources` becomes a separate Kubernetes resource in the output.

## Configuration Flow

Values flow from the bundle to the component automatically:

```cue
// In bundle.cue
odin.#Bundle & {
    components: {
        webapp: odin.#Component & {
            config: {
                image: string
                replicas: int | *1
            }
            resources: deployment: {
                // Uses config.image and config.replicas
            }
        }
    }
    
    values: {
        components: webapp: {
            image: "nginx:latest"
            replicas: 3
        }
    }
}
```

The `values.components.webapp` automatically unifies with `components.webapp.config`. This bidirectional binding is handled by Odin's internal schema.

## Conditional Resources

You can conditionally include resources using CUE's `if` statement:

```cue
resources: {
    deployment: {...}
    
    // Only create service if enabled
    if config.service.enabled {
        service: {
            apiVersion: "v1"
            kind: "Service"
            // ...
        }
    }
    
    // Create ingress only in production
    if config.environment == "production" {
        ingress: {
            apiVersion: "networking.k8s.io/v1"
            kind: "Ingress"
            // ...
        }
    }
}
```

## Using CUE Features

Components can leverage CUE's powerful features:

### List Comprehensions

```cue
// Convert map to list of env vars
env: [
    for k, v in config.env {
        name: k
        value: v
    }
]
```

### String Interpolation

```cue
image: "\(config.image.repository):\(config.image.tag)"
```

### References

```cue
resources: {
    deployment: {
        metadata: name: "myapp"
        spec: {
            selector: matchLabels: app: "myapp"
            template: {
                // Reference the selector labels
                metadata: labels: deployment.spec.selector.matchLabels
            }
        }
    }
}
```

### Definitions for Reuse

```cue
// Define common labels
#commonLabels: {
    app: metadata.name
    version: config.version
    team: "platform"
}

resources: {
    deployment: {
        metadata: labels: #commonLabels
        spec: template: metadata: labels: #commonLabels
    }
    service: {
        metadata: labels: #commonLabels
        spec: selector: #commonLabels
    }
}
```

## Component Organization

You can organize components in several ways:

### Inline in Bundle

```cue
odin.#Bundle & {
    components: {
        webapp: odin.#Component & {
            // Component definition here
        }
    }
}
```

### Separate Files

```cue
// components/webapp.cue
package components

import odin "go-valkyrie.com/odin/api/v1alpha1"

webapp: odin.#Component & {
    // Component definition
}
```

```cue
// bundle.cue
package main

import (
    odin "go-valkyrie.com/odin/api/v1alpha1"
    "example.com/myapp/components"
)

odin.#Bundle & {
    // Include all components from the package
    components
}
```

### Definitions for Reuse

```cue
// lib/webapp.cue
package lib

import odin "go-valkyrie.com/odin/api/v1alpha1"

#WebApp: odin.#Component & {
    config: {
        image: string
        replicas: int | *1
    }
    resources: {...}
}
```

```cue
// bundle.cue
import "example.com/myapp/lib"

odin.#Bundle & {
    components: {
        app1: lib.#WebApp & {
            metadata: name: "app1"
        }
        app2: lib.#WebApp & {
            metadata: name: "app2"
        }
    }
}
```

## Multiple Instances

You can create multiple instances of the same component type:

```cue
odin.#Bundle & {
    components: {
        webapp1: webapp.#Component & {
            metadata: name: "webapp1"
        }
        webapp2: webapp.#Component & {
            metadata: name: "webapp2"
        }
    }
    
    values: {
        components: {
            webapp1: {
                image: "app1:v1.0"
                replicas: 2
            }
            webapp2: {
                image: "app2:v2.0"
                replicas: 3
            }
        }
    }
}
```

## Next Steps

- Learn about [Resources](./resources.md) - the Kubernetes manifests components generate
- Understand [Component Templates](./templates.md) - reusable component definitions
- Explore [Writing Components](../guides/writing-components.md) - best practices and patterns
