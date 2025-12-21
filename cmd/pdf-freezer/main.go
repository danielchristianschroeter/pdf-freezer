package main

import (
	"log"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"

	pdffreezer "pdf-freezer"
	"pdf-freezer/pkg/app"
)

func main() {
	// Initialize the app service
	appService := app.NewApp()

	// Create application with options
	wailsApp := application.New(application.Options{
		Name:        "pdf-freezer",
		Description: "PDF Freezer Tool",
		Services: []application.Service{
			application.NewService(appService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(pdffreezer.Assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// Create main window with Drag & Drop enabled
	_ = wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:             "PDF Freezer",
		Width:             500,
		Height:            580,
		URL:               "/",
		BackgroundColour:  application.NewRGB(18, 18, 24),
		EnableDragAndDrop: true,
	})

	// Listen for the common:WindowFilesDropped event and re-emit to frontend
	wailsApp.Event.On("common:WindowFilesDropped", func(e *application.CustomEvent) {
		log.Printf("common:WindowFilesDropped received: %+v", e)
		handleFileDrop(wailsApp, e.Data)
	})

	err := wailsApp.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func handleFileDrop(app *application.App, data any) {
	var paths []string

	// Try to extract paths from the event data
	switch v := data.(type) {
	case map[string]any:
		if p, ok := v["filenames"].([]any); ok {
			for _, f := range p {
				if s, ok := f.(string); ok {
					paths = append(paths, s)
				}
			}
		} else if p, ok := v["paths"].([]any); ok {
			for _, f := range p {
				if s, ok := f.(string); ok {
					paths = append(paths, s)
				}
			}
		}
	case []any:
		for _, f := range v {
			if s, ok := f.(string); ok {
				paths = append(paths, s)
			}
		}
	case []string:
		paths = v
	}

	// Filter for PDFs
	var pdfPaths []string
	for _, p := range paths {
		if strings.HasSuffix(strings.ToLower(p), ".pdf") {
			pdfPaths = append(pdfPaths, p)
		}
	}

	if len(pdfPaths) > 0 {
		log.Printf("Emitting files-dropped with %d PDFs: %v", len(pdfPaths), pdfPaths)
		app.Event.Emit("files-dropped", pdfPaths)
	}
}
