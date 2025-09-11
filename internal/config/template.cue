package config

// Global CUE settings that apply to all Odin bundles
cue: {
	// A mapping of CUE module prefixes to OCI registries where said modules are stored, where the key is the CUE module
	// prefix and the value is the OCI registry to use.
	//
	// When the CUE loader is fetching modules, it will look for the longest matching prefix and use the associated
	// registry. The defaults here mean every module with a path starting with go-valkyrie.com or platform.go-valkyrie.com
	// will be loaded from the GitHub Container Registry.
	//
	// The registry specifier itself should be in the following format:
	//   hostname[:port][/repoPrefix][+insecure]
	//
	// The repoPrefix is optional, and is used to specify a prefix for the repository path within the registry, as done
	// with the defaults below.
	//
	// Adding +insecure at the end of the registry specifier will disable TLS for that registry, and the OCI client will
	// communicate with the registry over HTTP instead of HTTPS.
	registries: {
		"go-valkyrie.com": "ghcr.io/go-valkyrie/cue"
		"platform.go-valkyrie.com": "ghcr.io/go-valkyrie/cue"
	}
}
defaults: {
	prompt: false
}
