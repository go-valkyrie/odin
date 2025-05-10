package config

import (
	"net"
	"regexp"
	"strconv"
)

#modulePath: =~#"^((?:(?:\w|\d|-)+\.)+(?:\w+))((?:\/(?:\w|\d|-|_)+)*)$"#

#registryHost: S={
	string
	let parts = regexp.FindSubmatch(#"^([a-zA-Z0-9:.\[\]]+)((?:\/[a-z0-9\-_]+)+)?(\+insecure)?$"#, S)
	let host = parts[1]
	let path = parts[2]
	let insecure = parts[3]

	#host: string
	#path: string
	#port: uint

	if insecure == "" {
		#port: _ | *443
	}
	if insecure != "" {
		#port: _ | *80
	}

	#path: "" | =~#"^(\/[a-z0-9\-_]+)+(\+insecure)?$"#
	#path: path

	if net.SplitHostPort(host) != _|_ {
		let p = net.SplitHostPort(host)
		#host: p[0]
		#port: strconv.Atoi(p[1])
	}

	if net.SplitHostPort(host) == _|_ {
		#host: host
	}

	#host: =~ #"^((?:(?:\w|\d|-)+\.)(?:\w+))(:\d{1,5})?((?:\/(?:\w|\d|-|_)+)*)(\+insecure)?$"# | net.IP
}

#registries: {
	[#modulePath]: #registryHost
}

#cue: {
	registries: #registries
}

#config: {
	shared: string | *null
}

#ci: {
	provider: string
}

#defaults: {
	prompt: bool
}

cue: #cue
config: #config
defaults: #defaults

