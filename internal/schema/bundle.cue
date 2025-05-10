#Bundle: {
	apiVersion: "odin.go-valkyrie.com/v1alpha1"
	kind:       "Bundle"
	metadata: {
		name: string
		...
	}
	C=components: [Name=string]: {
		config: values.components[Name]
		...
	}
	values: {
		components: [string]: {...}
		components: {
			for name, _ in C {
				"\(name)": {...}
			}
		}
		...
	}
}
