package v1alpha1

_#TypeMeta: {
	apiVersion: string
	kind:       string
}

_#ObjectMeta: {
	name: string
}

#Bundle: {
	_#TypeMeta
	apiVersion: "odin.go-valkyrie.com/v1alpha1"
	kind:       "Bundle"
	metadata:   _#ObjectMeta
	components: [Name=string]: #ComponentBase
	values: {
		components: [string]: {...}
		...
	}
}
