package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var Dist embed.FS

// DistFS returns the embedded frontend build rooted at the "dist" directory.
func DistFS() (fs.FS, error) {
	return fs.Sub(Dist, "dist")
}
