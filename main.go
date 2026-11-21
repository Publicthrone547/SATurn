package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/icons"
)

//go:embed all:frontend/dist
var assets embed.FS

// version is set at build time (-ldflags "-X main.version=...").
var version = "0.0.0-dev"

const bundleID = "io.github.publicthrone547.saturn"

func main() {
	app := application.New(application.Options{
		Name:        "SATurn",
		Description: "SAT study reminder",
		Assets:      application.AssetOptions{Handler: application.AssetFileServerFS(assets)},
		Mac: application.MacOptions{
			ActivationPolicy: application.ActivationPolicyAccessory,
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
		SingleInstance: &application.SingleInstanceOptions{UniqueID: bundleID},
	})
	popover := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name: "popover", URL: "/#/popover", Width: 360, Height: 540,
		Frameless: true, AlwaysOnTop: true, Hidden: true, DisableResize: true,
		HideOnEscape: true, HideOnFocusLost: true,
		BackgroundColour: application.NewRGB(13, 16, 48),
	})
	tray := app.SystemTray.New()
	tray.SetTemplateIcon(icons.SystrayMacTemplate)
	tray.SetLabel("SATurn")
	menu := application.NewMenu()
	menu.Add("Quit SATurn").OnClick(func(*application.Context) { app.Quit() })
	tray.SetMenu(menu)
	tray.AttachWindow(popover).WindowOffset(6)
	tray.OnRightClick(func() { tray.OpenMenu() })
	log.Printf("start version=%s", version)
	log.Printf("tray created")
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
