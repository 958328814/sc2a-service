package ui

import (
	"log"
	"net/http"

	_ "github.com/DreamHacks/sc2a-service/ui/statik"
	"github.com/rakyll/statik/fs"
)

//go:generate statik -src=./build

// FS is a http.FileSystem to serve packed assets
var FS http.FileSystem

func init() {
	var err error
	FS, err = fs.New()
	if err != nil {
		log.Fatal(err)
	}
}
