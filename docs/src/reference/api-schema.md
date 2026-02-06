# CUE API Schema

This reference documents Odin's CUE API schemas.

## Package: go-valkyrie.com/odin/api/v1alpha1

### #Bundle

The top-level bundle schema.

```cue
#Bundle: {
    apiVersion: "odin.go-valkyrie.com/v1alpha1"
    kind: "Bundle"
    
    // Bundle metadata
    metadata: {
        // Required: bundle name
        name: string
    }
    
    // Components map
    components: [Name=string]: #ComponentBase
    
    // Configuration values
    values: {
        // Component-specific values
        components: [string]: {...}
        // Additional fields allowed
        ...
    }
}
```

**Fields:**

- `metadata.name` (string, required) - Bundle name
- `components` (map, required) - Map of component name to component definition
- `values` (struct, optional) - Configuration values for components

### #Component

Standard component definition extending `#ComponentBase`.

```cue
#Component: #ComponentBase & {
    // Component metadata
    metadata: {
        // Required: component name
        name: string
        
        // Optional labels
        labels?: [string]: string
        
        // Optional annotations
        annotations?: [string]: string
    }
    
    // Configuration schema
    config: {
        // Define your configuration fields here
        ...
    }
    
    // Kubernetes resources
    resources: {
        [string]: {
            // Kubernetes resource definition
            apiVersion: string
            kind: string
            metadata?: {...}
            spec?: {...}
            data?: {...}
            ...
        }
    }
}
```

**Fields:**

- `metadata.name` (string, required) - Component name
- `metadata.labels` (map, optional) - Labels for all resources
- `metadata.annotations` (map, optional) - Annotations for all resources
- `config` (struct, required) - Configuration schema
- `resources` (map, required) - Kubernetes resources to generate

### #ComponentBase

Base interface that all components must implement.

```cue
#ComponentBase: {
    metadata: {
        name: string
        labels?: [string]: string
        annotations?: [string]: string
    }
    config: {...}
    resources: [string]: {...}
}
```

All custom component definitions should extend `#ComponentBase`.

## Type Reference

### Metadata Types

**_#TypeMeta:**

```cue
_#TypeMeta: {
    apiVersion: string
    kind: string
}
```

**_#ObjectMeta:**

```cue
_#ObjectMeta: {
    name: string
}
```

## Creating Custom Component Types

Extend `#ComponentBase` to create reusable component templates:

```cue
package v1alpha1

import odin "go-valkyrie.com/odin/api/v1alpha1"

#WebApp: odin.#ComponentBase & {
    config: {
        image: string
        replicas: int | *1
    }
    
    resources: {
        deployment: {...}
        service: {...}
    }
}
```

## Schema Validation

The API schema provides compile-time validation:

```cue
// Valid: correct structure
odin.#Bundle & {
    metadata: name: "my-app"
    components: {...}
    values: {...}
}

// Error: missing required field
odin.#Bundle & {
    components: {...}
    // Error: field metadata not found
}

// Error: wrong type
odin.#Bundle & {
    metadata: name: 123
    // Error: conflicting values 123 and string
}
```

## Importing the API

```cue
package main

import (
    odin "go-valkyrie.com/odin/api/v1alpha1"
)

// Use the schema
odin.#Bundle & {
    metadata: name: "my-bundle"
    // ...
}
```

## Version Compatibility

The API follows semantic versioning:

- `v1alpha1@v0` - Development, breaking changes allowed
- `v1@v1` - Stable v1 (future)

Import the latest v0 release:

```cue
deps: {
    "go-valkyrie.com/odin/api/v1alpha1@v0": {
        v: "v0.3.0"
    }
}
```

## Next Steps

- See [Configuration Format](./config-format.md) for Odin configuration
- Learn [Best Practices](./best-practices.md) for using the API
- Explore [Writing Components](../guides/writing-components.md)
