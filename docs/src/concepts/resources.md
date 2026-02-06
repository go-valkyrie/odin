# Resources

**Resources** are the Kubernetes manifests that components generate. Each entry in a component's `resources` section becomes a separate YAML document in the output.

## What is a Resource?

A resource is:

- A Kubernetes API object (Deployment, Service, ConfigMap, etc.)
- Defined as a field in a component's `resources` section
- Rendered as YAML by the `odin template` command
- Separated by `---` in multi-document YAML output

## Resource Structure

Resources are plain Kubernetes manifests written in CUE:

```cue
resources: {
    deployment: {
        apiVersion: "apps/v1"
        kind: "Deployment"
        metadata: {
            name: "myapp"
            namespace: "default"
            labels: {
                app: "myapp"
            }
        }
        spec: {
            replicas: 3
            selector: {
                matchLabels: app: "myapp"
            }
            template: {
                metadata: labels: app: "myapp"
                spec: containers: [{
                    name: "myapp"
                    image: "nginx:latest"
                    ports: [{
                        containerPort: 80
                    }]
                }]
            }
        }
    }
}
```

## Resource Naming

The key in the `resources` map is the resource name used for:

- Identifying the resource in Odin's output comments
- Organizing resources within a component
- Conditional resource inclusion

**Important:** The resource key is NOT the Kubernetes resource name. That's defined in `metadata.name`.

```cue
resources: {
    // "main-deployment" is the Odin resource name
    "main-deployment": {
        kind: "Deployment"
        metadata: {
            // "webapp" is the Kubernetes resource name
            name: "webapp"
        }
    }
}
```

## Resource Output Format

When you run `odin template`, resources are rendered as YAML:

```yaml
---
# Source: mycomponent/deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 3
  # ...

---
# Source: mycomponent/service
apiVersion: v1
kind: Service
metadata:
  name: myapp
spec:
  # ...
```

The `# Source: component/resource` comment shows where each resource came from.

## Multiple Resources per Component

Components commonly generate multiple related resources:

```cue
resources: {
    deployment: {
        apiVersion: "apps/v1"
        kind: "Deployment"
        // ...
    }
    
    service: {
        apiVersion: "v1"
        kind: "Service"
        // ...
    }
    
    configmap: {
        apiVersion: "v1"
        kind: "ConfigMap"
        metadata: name: "myapp-config"
        data: {
            "config.yaml": yaml.Marshal(config.appConfig)
        }
    }
    
    serviceaccount: {
        apiVersion: "v1"
        kind: "ServiceAccount"
        metadata: name: "myapp"
    }
}
```

## Conditional Resources

Use CUE's `if` statement to conditionally include resources:

```cue
config: {
    ingress: {
        enabled: bool | *false
        hostname: string
    }
}

resources: {
    deployment: {...}
    service: {...}
    
    // Only create ingress if enabled
    if config.ingress.enabled {
        ingress: {
            apiVersion: "networking.k8s.io/v1"
            kind: "Ingress"
            metadata: name: metadata.name
            spec: rules: [{
                host: config.ingress.hostname
                http: paths: [{
                    path: "/"
                    pathType: "Prefix"
                    backend: service: {
                        name: metadata.name
                        port: number: 80
                    }
                }]
            }]
        }
    }
}
```

## Resource References

Resources can reference each other within the same component:

```cue
resources: {
    deployment: {
        metadata: name: "myapp"
        spec: {
            template: spec: {
                serviceAccountName: resources.serviceaccount.metadata.name
                volumes: [{
                    name: "config"
                    configMap: {
                        name: resources.configmap.metadata.name
                    }
                }]
            }
        }
    }
    
    serviceaccount: {
        metadata: name: "myapp-sa"
    }
    
    configmap: {
        metadata: name: "myapp-config"
    }
}
```

## Dynamic Resource Generation

You can generate resources dynamically using CUE's list comprehensions:

```cue
config: {
    services: {
        web: port: 8080
        api: port: 8081
        admin: port: 8082
    }
}

resources: {
    // Generate a service for each configured service
    for name, svc in config.services {
        "service-\(name)": {
            apiVersion: "v1"
            kind: "Service"
            metadata: {
                name: "\(metadata.name)-\(name)"
            }
            spec: {
                selector: app: metadata.name
                ports: [{
                    port: svc.port
                    targetPort: svc.port
                }]
            }
        }
    }
}
```

This generates three separate services: `service-web`, `service-api`, and `service-admin`.

## Resource Validation

CUE provides validation at multiple levels:

### Type Checking

```cue
resources: {
    deployment: {
        apiVersion: "apps/v1"
        kind: "Deployment"
        spec: {
            replicas: "three"  // Error: string not allowed for int field
        }
    }
}
```

### Required Fields

```cue
resources: {
    deployment: {
        apiVersion: "apps/v1"
        kind: "Deployment"
        // Error: missing required field 'metadata'
        spec: {...}
    }
}
```

### Constraints

```cue
config: {
    replicas: int & >=1 & <=10  // Must be between 1 and 10
}

resources: {
    deployment: {
        spec: replicas: config.replicas
    }
}
```

## Using Kubernetes CUE Schemas

You can import Kubernetes API schemas for full validation:

```cue
import (
    apps "k8s.io/api/apps/v1"
    core "k8s.io/api/core/v1"
)

resources: {
    deployment: apps.#Deployment & {
        apiVersion: "apps/v1"
        kind: "Deployment"
        // ... 
    }
    
    service: core.#Service & {
        apiVersion: "v1"
        kind: "Service"
        // ...
    }
}
```

This provides IDE completion and full API validation.

## Resource Ordering

Odin outputs resources in a deterministic order:

1. Resources are ordered by component name (alphabetically)
2. Within a component, resources are ordered by their key (alphabetically)

If you need specific ordering for `kubectl apply`, use multiple `odin template` commands or tools like Kustomize.

## Complex Resources

For complex resources, you can use definitions to break down the structure:

```cue
#container: {
    name: string
    image: string
    port: int
    
    // Full container definition
    _out: {
        name: name
        image: image
        ports: [{containerPort: port}]
    }
}

config: {
    containers: {
        web: #container & {
            name: "web"
            image: "nginx:latest"
            port: 8080
        }
        sidecar: #container & {
            name: "envoy"
            image: "envoyproxy/envoy:latest"
            port: 9901
        }
    }
}

resources: {
    deployment: {
        spec: template: spec: {
            containers: [
                for _, c in config.containers {
                    c._out
                }
            ]
        }
    }
}
```

## Next Steps

- Learn about [Values](./values.md) - how configuration flows into resources
- Explore [Writing Components](../guides/writing-components.md) - patterns for resource generation
- See [Best Practices](../reference/best-practices.md) - tips for managing resources
