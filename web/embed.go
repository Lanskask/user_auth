// Package web embeds the built frontend (web/static) into the Go binary.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:static
var static embed.FS

// Static returns the built frontend rooted at its index.html.
func Static() (fs.FS, error) { return fs.Sub(static, "static") }
