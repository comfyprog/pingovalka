package frontend

import (
	"embed"
	"io/fs"
)

//go:embed dist
var distFs embed.FS
var FrontendFs fs.FS

func init() {
	var err error
	FrontendFs, err = fs.Sub(distFs, "dist")
	if err != nil {
		panic(err)
	}
}
