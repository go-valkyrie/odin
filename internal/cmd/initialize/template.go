// SPDX-License-Identifier: MIT

package initialize

import (
	"embed"
	"text/template"
)

//go:embed template/*
var templateFS embed.FS

var bundleTemplate = func() *template.Template {
	return template.Must(template.ParseFS(templateFS, "template/*"))
}()
