package main

import (
	"embed"
	"io/fs"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	sub, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}

	app := NewApp()

	err = wails.Run(&options.App{
		Title:  "网络代理",
		Width:  460,
		Height: 400,
		AssetServer: &assetserver.Options{
			Assets: sub,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
