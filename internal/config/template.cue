// Thanks for using Valkyrie! We've provided this default configuration file to help you get started. In addition to
// providing default settings,there's also a complete skeleton of the available options and comments to describe how to
// use them (and why you might want to).
//
// Since the config is written in CUE, you can use all the same techniques you would apply to your Valkyrie project
// definitons to do things like template boilerplate configuration, etc. Your config is validated against the CUE schema
// embedded in the Valkyrie CLI, if you want to see the definitions for the version of Valkyrie you've installed just
// run `valkyrie config schema` (this is a hidden command, it does not show up in the help output). You can also get an
// updated version of this default configuration file by running `valkyrie config defaults`.
//
// This file will never be automatically modified by Valkyrie, so if breaking changes are made to the schema in future
// versions of Valkyrie you will need to manually update this file (or delete it and the current defaults will be placed
// here again). Should this occur the release notes will contain instructions on how to update your config.
//
// A quick note before we get started, since this file also configures registries that Valkyrie will use to fetch CUE
// modules, you cannot use the registries defined here to import modules into this config file itself. The most common
// use case for doing this would be importing a shared configuration from a registry, see the `config` section below
// where there's explicit support for doing that.

package config

// Global CUE settings that apply to all Valkyrie projects
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

config: {
	// The package specifier for a shared configuration module provided by your Platform Adiminstrator. If this is set,
	// you can use `valkyrie config pull` to fetch the latest version of the module and the settings in it will be unified
	// with your local config.
	//
	// This should be in the form of a full CUE package specifier, e.g.:
	//   "platform.company.com/config[@vN]"
	//
	// Do note that the standard rules of CUE unification apply, so if there are conflicts between your local config and
	// that from the shared module you will get an error message when running the Valkyrie CLI, and will either need to:
	//
	// 1. Remove or comment outthe conflicting configuration from this file,
	// 2. Comment out the below line to disable using the shared module, you can always merge the settings you actually
	//    need or want into your local config.
	// 3. Email your Platform Administrator and let them know the latest update to the shared config is conflicting with
	//    your local config, and they should stop breaking your config. (Your mileage may vary)
	shared: null
}

// Defaults that are used when initializing Valkyrie for a project.
defaults: {
	// This determines whether Valkyrie will prompt you to confirm you want to use these settings when configuring a
	// project for the first time by running `valkyrie init`. If you're in a situation where you need to work with multiple
	// platform definitions, CI providers, etc. regularly you can either set this to `true` to be prompted to confirm all
	// the settings when initializing a project, or specify use the appropriate flags when calling `valkyrie init` (these
	// will take precedence over the settings here).
	prompt: false
	// The default platform configuration to use. To get you started we provide the standard platform definition provided
	// by the Valkyrie team, but this only covers the most common use cases that can be supported by Valkyrie itself. If
	// you're using Valkyrie for anything but personal projects, you probably want to use a custom platform definition to
	// provide additional functionality, such as additional component types, traits, etc.
	platform: "platform.go-valkyrie.com/platform/v1alpha1"
	ci: {
		// The default CI provider to use for projects. When you run `valkyrie init` to setup Valkyrie for a project, thi
		provider: "github"
	}
}
