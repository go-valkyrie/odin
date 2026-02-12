// SPDX-License-Identifier: MIT
package workload

import odin "go-valkyrie.com/odin/api/v1alpha1"

// #WebApp is a Deployment with a Service for testing
#WebApp: C=odin.#Component & {
	_config: config

	config: {
		image:    string
		replicas: uint
		port:     uint | *80
	}

	resources: {
		deployment: {
			apiVersion: "apps/v1"
			kind:       "Deployment"
			metadata: name: C.metadata.name
			spec: {
				replicas: _config.replicas
				selector: matchLabels: app: C.metadata.name
				template: {
					metadata: labels: app: C.metadata.name
					spec: containers: [{
						name:  C.metadata.name
						image: _config.image
						ports: [{
							containerPort: _config.port
						}]
					}]
				}
			}
		}

		service: {
			apiVersion: "v1"
			kind:       "Service"
			metadata: name: C.metadata.name
			spec: {
				selector: app: C.metadata.name
				ports: [{
					port:       _config.port
					targetPort: _config.port
				}]
			}
		}
	}
}
