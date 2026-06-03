package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	auditService := NewAuditService()

	err := wails.Run(&options.App{
		Title:    "Spreadsheet Auditor",
		Width:    1200,
		Height:   800,
		MinWidth: 1100,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 248, G: 249, B: 250, A: 1},
		OnStartup:        auditService.startup,
		Bind: []interface{}{
			auditService,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
