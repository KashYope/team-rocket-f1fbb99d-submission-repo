package main

import (
	"embed"
	"log"
	"time"

	"es-incident-command/internal/service"

	"github.com/joho/godotenv"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Load .env if present; ignore error in pure zero-config offline mode
	_ = godotenv.Load()

	incidentSvc := service.NewIncidentService()

	app := application.New(application.Options{
		Name:        "ES Incident Command",
		Description: "European Sleeper Night Train Disruption Command Center PoC",
		Services: []application.Service{
			application.NewService(incidentSvc),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Server: application.ServerOptions{
			Host:         "localhost",
			Port:         8080,
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "ES Incident Command — Night Train Ops PoC",
		Width:            1440,
		Height:           900,
		MinWidth:         1180,
		MinHeight:        760,
		BackgroundColour: application.NewRGB(15, 23, 42),
		URL:              "/",
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
