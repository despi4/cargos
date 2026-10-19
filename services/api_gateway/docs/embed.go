// Package docs bundles the OpenAPI contract and a local Swagger UI.
package docs

import (
	"embed"
	"io/fs"
)

//go:embed openapi.json
var Spec []byte

//go:embed swagger/*
var files embed.FS
var Files fs.FS

func init() {
	var err error
	Files, err = fs.Sub(files, "swagger")
	if err != nil {
		panic(err)
	}
}
