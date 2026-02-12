// SPDX-License-Identifier: MIT
package workload

import odin "go-valkyrie.com/odin/api/v1alpha1"

// #Deployment is a simple Kubernetes Deployment template for testing
#Deployment: C=odin.#Component & {
	_config: config

	config: {
		image:    string
		replicas: uint
	}

	resources: deployment: {
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
				}]
			}
		}
	}
}
