![logo](./docs/logo.png)

# Generate Kubernetes manifests with CUE

Odin is a tool to generate Kubernetes manifests using the CUE language. It is
not a package manager, it does not know how to apply resources to a Kubernetes
cluster, it *just* generates manifests.

While designed to be used alongside [Valkyrie](https://github.com/go-valkyrie/valkyrie)
and [Freyr](https://github.com/go-valkyrie/freyr), Odin is completely usable as
a component of any GitOps pipeline that uses the rendered manifest pattern (as is
done with Freyr).

## Why should I use Odin? (Elevator Pitch)

Existing tools for templating Kubernetes manifests fall into multiple traps:

* Using brittle text-based templates (Helm)
* Do not integrate nicely with external tools like ArgoCD (Timoni, KubeVela)
* Make easy-to-use abstractions difficult (KCL, Pkl, ytt)

All the tools listed above are entirely valid choices, to be clear, but Odin was
designed with the following goals:

* Be simple enough for a junior developer to pick up in an hour
* Support deploying arbitrary Kubernetes resources, while allowing the creation
  of simple abstractions for common use cases
* Allow 'build-time' validation of configurations (avoid errors at apply-time)
* Integrate nicely with ArgoCD (or any other CD tool of your choice that)

## A simple Odin bundle

An Odin 'bundle' consists of a set of components, their resources, and a set of
configurable values that can be used to customize the rendered manifests. You can
see a very simple bundle below that simply defines (most of) a Kubernetes Deployment
and Service, and will result in exactly the same YAML you would expect.

```cue
package example

import (
    odin "go-valkyrie.com/odin/api/v1alpha1"
)

odin.#Bundle & {
    metadata: {
        name: "example"
    }
    components: {
        httpbin: odin.#Component & {
            metadata: name: "httpbin"
            config: {
                image: string
            }
            resources: deployment: {
                apiVersion: "apps/v1"
                kind: "Deployment"
                metadata: name: "httpbin"
                spec: {
                    template: spec: containers: [{
                        name: "httpbin"
                        image: config.image
                        ...
                    }]
                }
            }
            resources: service: {
                apiVersion: "v1"
                kind: "Service"
                ...
            }
        ...
        }
    }
    values: {
        components: httpbin: image: "kennethreitz/httpbin:latest"
    }
}
```

# Building abstractions

Writing raw Kubernetes manifests in an Odin bundle is fine, but bundles being
written in CUE allows us to define abstractions for common use cases. In fact,
the above example could be written much more succinctly using the `#WebApp`
definition provided by Freyr (do note, all the Odin component definitions provided
by the Freyr project can be used independently) like so:

```cue
package example

import (
    odin "go-valkyrie.com/odin/api/v1alpha1"
    webapp "platform.go-valkyrie.com/webapp/v1alpha1"
)

odin.#Bundle & {
	metadata: {
		name: "example"
	}
	components: {
		httpbin: webapp.#Component & {
			metadata: name: "httpbin"
		}
	}
	values: {
		components: httpbin: {
			image: name: "kennethreitz/httpbin"
			image: tag: "latest"
		}
	}
}
```

The component definitions provided by Freyr are also extensible, you can
add an ingress to the above making a small change to the above example:

```cue
components: {
	httpbin: webapp.#Component & {
		traits: [
			webapp.#Ingress
		]
	}
}
values: {
    components: httpbin: {
        ingress: hostname: "example.com"
    }
}
```

Building these abstractions is no harder than writing standard Odin
component definitions, since that's all they are. This also provides a
way to share common functionality between multiple application by packaging
it up into a CUE module and publishing it to an OCI registry, so if you have
a standard template for deploying an OpenTelemetry collector, a Redis instance,
etc. you can define it once and pull it into every bundle that deploys those
services (this is also why Odin explicitly does not support something like
Helm's subchart functionality, as bundles can be composed of components from
multiple CUE modules).