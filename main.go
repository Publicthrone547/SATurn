// Command SATurn is a menu bar study reminder: one ring a day, closed by
// logging the day's questions, with hourly reminders until it is closed.
package main

import (
	"embed"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
)

//go:embed all:frontend/dist
var assets embed.FS

// version is set at build time (-ldflags "-X main.version=...").
var version = "0.0.0-dev"

const bundleID = "io.github.publicthrone547.saturn"

func main() {
	setupLog()
	e := readEnv()
	notes := notifications.New()
	sh, err := newShell(notes, e)
	if err != nil {
		log.Fatalf("state: %v", err)
	}
	svc := &AppService{sh: sh}

	app := application.New(application.Options{
		Name:        "SATurn",
		Description: "SAT study reminder",
		Services: []application.Service{
			application.NewService(svc),
			application.NewService(notes),
		},
		Assets: application.AssetOptions{Handler: application.AssetFileServerFS(assets)},
		Mac: application.MacOptions{
			ActivationPolicy: application.ActivationPolicyAccessory,
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID:               bundleID,
			OnSecondInstanceLaunch: func(application.SecondInstanceData) { sh.openMain("") },
		},
		ShouldQuit: sh.shouldQuit,
		OnShutdown: sh.shutdown,
	})
	sh.app = app

	sh.popover = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name: "popover", URL: "/#/popover", Width: 360, Height: 540,
		Frameless: true, AlwaysOnTop: true, Hidden: true, DisableResize: true,
		HideOnEscape: true, HideOnFocusLost: true,
		BackgroundColour: application.NewRGB(13, 16, 48),
	})
	sh.main = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name: "main", Title: "SATurn", URL: "/#/main",
		Width: 920, Height: 640, MinWidth: 760, MinHeight: 520, Hidden: true,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 38,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(13, 16, 48),
	})
	// Closing the main window hides it; the app lives in the menu bar.
	sh.main.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		sh.main.Hide()
		e.Cancel()
	})
	sh.popover.OnWindowEvent(events.Common.WindowFocus, func(*application.WindowEvent) { go sh.popoverShown() })
	sh.island = app.Window.NewNotchWindow(application.NotchWindowOptions{
		Width: 560, Height: 104, Animated: true,
		WindowOptions: application.WebviewWindowOptions{Name: "island", URL: "/#/island", Hidden: true},
	})

	sh.tray = app.SystemTray.New()
	sh.buildTray()

	notes.OnNotificationResponse(sh.onNotificationResponse)
	app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) { go sh.start() })
	app.Event.OnApplicationEvent(events.Common.SystemDidWake, func(*application.ApplicationEvent) { go sh.woken() })
	app.Event.OnApplicationEvent(events.Mac.ApplicationShouldHandleReopen, func(*application.ApplicationEvent) { go sh.openMain("") })

	// A signal from outside (kill, pkill, launchd) always ends the app, strict
	// mode or not; the state is saved after every change anyway.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	go func() {
		s := <-sig
		log.Printf("signal %v", s)
		sh.allowQuit.Store(true)
		sh.shutdown()
		go func() {
			time.Sleep(3 * time.Second)
			os.Exit(0)
		}()
		sh.quit()
	}()

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
