// SPDX-License-Identifier: MIT

package cmd

// RunningEmbedded is a simple protection mechanism to prevent the entrypoint to the `odin` CLI from executing
// tasks that are unsafe to use when it is being called as a Go library
var RunningEmbedded = true
