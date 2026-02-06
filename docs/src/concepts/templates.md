# Component Templates

**Component templates** are reusable component definitions that can be imported from CUE modules. They allow you to share common patterns across multiple bundles and teams.

## What are Component Templates?

Component templates are:

- CUE definitions that extend `#ComponentBase`
- Published as CUE modules to OCI registries
- Imported into bundles like any CUE module
- Configurable through their `config` schema

## Using Templates

### Import a Template

First, add the module to your `cue.mod/module.cue`:

```cue
module: "example.com/my-bundle@v0"
language: version: "v0.11.0"

deps: {
    "platform.go-valkyrie.com/webapp/v1alpha1@v0": {
        v: "v0.1.0"
    }
}
```

Then import and use it in your bundle:

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
            image: name: "myapp"
            image: tag: "v1.0.0"
        }
    }
}
```

### Discover Available Templates

List templates in your bundle's dependencies:

```bash
odin components
```

Output:

```
PACKAGE                                              NAME         MODULE                                VERSION
platform.go-valkyrie.com/webapp/v1alpha1             #Component   platform.go-valkyrie.com/webapp       v0.1.0
platform.go-valkyrie.com/database/v1alpha1           #Component   platform.go-valkyrie.com/database     v0.1.0
platform.go-valkyrie.com/gateway/v1alpha1            #Component   platform.go-valkyrie.com/gateway      v0.2.0
```

### View Template Documentation

See what configuration a template accepts:

```bash
odin docs webapp
```

Or with the full package name:

```bash
odin docs platform.go-valkyrie.com/webapp/v1alpha1
```

This shows the template's config schema, available fields, and documentation.

## Creating Templates

### Basic Template Structure

Create a CUE definition that extends `#ComponentBase`:

```cue
package v1alpha1

import (
    odin "go-valkyrie.com/odin/api/v1alpha1"
)

// #Component is a web application with a Deployment and Service
#Component: odin.#ComponentBase & {
    config: {
        // Image configuration
        image: {
            repository: string
            tag: string | *"latest"
            pullPolicy: string | *"IfNotPresent"
        }
        
        // Scaling
        replicas: int & >=1 & <=100 | *1
        
        // Networking
        port: int & >0 & <65536 | *8080
        
        // Resources
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
    }
    
    resources: {
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
                        image: "\(config.image.repository):\(config.image.tag)"
                        imagePullPolicy: config.image.pullPolicy
                        ports: [{
                            containerPort: config.port
                            name: "http"
                        }]
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
                    targetPort: "http"
                    name: "http"
                }]
            }
        }
    }
}
```

### Module Structure

Organize your template module:

```
webapp/
├── cue.mod/
│   └── module.cue
├── v1alpha1/
│   ├── component.cue      # Main component definition
│   ├── traits.cue         # Optional traits/mixins
│   └── doc.cue            # Documentation
└── examples/
    └── basic.cue
```

**cue.mod/module.cue:**

```cue
module: "platform.example.com/webapp@v0"
language: version: "v0.11.0"

deps: {
    "go-valkyrie.com/odin/api/v1alpha1@v0": {
        v: "v0.3.0"
    }
}
```

## Template Versioning

Templates follow semantic versioning in the module path:

```
platform.example.com/webapp/v1alpha1@v0    →  v0.x.x (breaking changes allowed)
platform.example.com/webapp/v1@v1          →  v1.x.x (stable API)
platform.example.com/webapp/v2@v2          →  v2.x.x (new major version)
```

Import specific versions:

```cue
deps: {
    "platform.example.com/webapp/v1alpha1@v0": {
        v: "v0.2.0"  // Pin to specific version
    }
}
```

## Template Composition

Templates can compose other templates:

```cue
package v1alpha1

import (
    odin "go-valkyrie.com/odin/api/v1alpha1"
    workload "platform.example.com/workload/v1"
)

#WebApp: workload.#Component & {
    // Add web-specific configuration
    config: {
        // Inherit workload config
        workload.#Component.config
        
        // Add ingress config
        ingress: {
            enabled: bool | *false
            hostname: string
            tls: bool | *false
        }
    }
    
    // Add ingress resource
    if config.ingress.enabled {
        resources: ingress: {
            apiVersion: "networking.k8s.io/v1"
            kind: "Ingress"
            // ...
        }
    }
}
```

## Traits and Mixins

Create optional features as traits:

```cue
package v1alpha1

// #Instrumentation adds OpenTelemetry instrumentation
#Instrumentation: {
    config: {
        instrumentation: {
            enabled: bool | *false
            collector: string | *"otel-collector:4317"
            serviceName: string
        }
    }
    
    if config.instrumentation.enabled {
        resources: deployment: {
            spec: template: {
                metadata: annotations: {
                    "instrumentation.opentelemetry.io/inject": "true"
                }
                spec: {
                    env: [
                        {
                            name: "OTEL_EXPORTER_OTLP_ENDPOINT"
                            value: config.instrumentation.collector
                        },
                        {
                            name: "OTEL_SERVICE_NAME"
                            value: config.instrumentation.serviceName
                        },
                    ]
                }
            }
        }
    }
}
```

Use traits with templates:

```cue
import (
    webapp "platform.example.com/webapp/v1alpha1"
)

components: {
    myapp: webapp.#Component & webapp.#Instrumentation & {
        metadata: name: "myapp"
    }
}

values: {
    components: myapp: {
        image: {...}
        instrumentation: {
            enabled: true
            serviceName: "myapp"
        }
    }
}
```

## Publishing Templates

### 1. Tag Your Release

```bash
git tag v0.1.0
git push origin v0.1.0
```

### 2. Push to OCI Registry

```bash
cue mod publish v0.1.0
```

### 3. Configure Registry

Add to your `odin.toml`:

```toml
[[registries]]
prefix = "platform.example.com"
registry = "ghcr.io/myorg/cue"
```

## Local Templates

For templates not yet ready to publish, use local imports:

```
my-bundle/
├── bundle.cue
├── lib/
│   └── webapp/
│       └── component.cue
└── cue.mod/
    └── module.cue
```

**lib/webapp/component.cue:**

```cue
package webapp

import odin "go-valkyrie.com/odin/api/v1alpha1"

#Component: odin.#ComponentBase & {
    // Template definition
}
```

**bundle.cue:**

```cue
package main

import (
    odin "go-valkyrie.com/odin/api/v1alpha1"
    webapp "example.com/my-bundle/lib/webapp"
)

odin.#Bundle & {
    components: {
        app: webapp.#Component & {
            metadata: name: "app"
        }
    }
}
```

## Template Best Practices

### 1. Document Your Configuration

```cue
#Component: {
    config: {
        // Number of pod replicas (1-100)
        // Default: 1
        replicas: int & >=1 & <=100 | *1
        
        // Container image configuration
        image: {
            // Image repository without tag
            // Example: "ghcr.io/myorg/myapp"
            repository: string
            
            // Image tag or digest
            // Default: "latest"
            tag: string | *"latest"
        }
    }
}
```

### 2. Provide Sensible Defaults

```cue
config: {
    replicas: int | *1
    port: int | *8080
    resources: {
        requests: {
            memory: string | *"128Mi"
            cpu: string | *"100m"
        }
    }
}
```

### 3. Validate Configuration

```cue
config: {
    replicas: int & >=1 & <=100
    port: int & >0 & <65536
    environment: "dev" | "staging" | "prod"
}
```

### 4. Make Features Optional

```cue
config: {
    ingress: {
        enabled: bool | *false
        ...
    }
    monitoring: {
        enabled: bool | *false
        ...
    }
}

if config.ingress.enabled {
    resources: ingress: {...}
}
```

### 5. Use Clear Names

```cue
// Good
#WebApp: {...}
#DatabaseCluster: {...}
#APIGateway: {...}

// Bad
#Comp1: {...}
#Thing: {...}
```

## Next Steps

- Learn [Using Component Templates](../guides/using-templates.md) in detail
- See [Publishing Components](../guides/publishing.md) to share your templates
- Explore [Configuration](../guides/configuration.md) for registry setup
