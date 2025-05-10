package v1alpha1

#ComponentBase: {
	_#TypeMeta
	apiVersion: string
	kind:       string
	metadata:   _#ObjectMeta
	config: {...}
	resources: [string]: {
		...
	}
}

#Component: #ComponentBase & {
	apiVersion: "odin.go-valkyrie.com/v1alpha1"
	kind:       "BasicComponent"
}
