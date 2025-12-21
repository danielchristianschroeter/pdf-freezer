package app

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"

	"pdf-freezer/internal/config"
	"pdf-freezer/internal/counter"
	"pdf-freezer/internal/engine"
)

// App struct
type App struct {
	counter  *counter.Manager
	config   *config.Manager
	logger   *config.Logger
	pipeline *engine.Pipeline
}

// NewApp creates a new App application struct
func NewApp() *App {
	// Init Logger
	l, err := config.NewLogger()
	if err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
	} else {
		l.Info("App starting...")
	}

	// Init Counter
	c, err := counter.NewManager()
	if err != nil {
		if l != nil {
			l.Error(fmt.Sprintf("Failed to init counter: %v", err))
		}
	}

	// Init Config
	cfg, err := config.NewManager()
	if err != nil {
		if l != nil {
			l.Error(fmt.Sprintf("Failed to init config: %v", err))
		}
	}

	// Initialize Pipeline
	p := engine.NewPipeline(c)

	return &App{
		counter:  c,
		config:   cfg,
		logger:   l,
		pipeline: p,
	}
}

// CheckDeps checks if system dependencies (GS) are met
func (a *App) CheckDeps() error {
	wrapper := engine.NewGhostscriptWrapper()
	err := wrapper.CheckDependencies()
	if err != nil {
		if a.logger != nil {
			a.logger.Error(fmt.Sprintf("CheckDeps failed: %v", err))
		}
	}
	return err
}

// OnFileDrop handles the file drop event
func (a *App) OnFileDrop(paths []string) {
	if len(paths) > 0 {
		path := paths[0]
		if a.logger != nil {
			a.logger.Info(fmt.Sprintf("File dropped: %s", path))
		}
		// Emit event using v3 API
		app := application.Get()
		if app != nil {
			app.Event.Emit("file-dropped", path)
		}
	}
}

// SelectFile opens a dialog to select a PDF
func (a *App) SelectFile() (string, error) {
	app := application.Get()

	dialog := app.Dialog.OpenFile()
	dialog.SetTitle("Select PDF to Freeze")
	dialog.AddFilter("PDF Files", "*.pdf")

	// Attach to current window if possible.
	// Current() returns Window interface.
	if currentWindow := app.Window.Current(); currentWindow != nil {
		dialog.AttachToWindow(currentWindow)
	}

	// PromptForSingleSelection returns (string, error)
	file, err := dialog.PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	return file, nil
}

// ProcessFile freezes the PDF
func (a *App) ProcessFile(inputPath string, overlayOverride bool, prefixOverride string, positionOverride string, suffixOverride string, overwriteMode bool, compressionLevel string) (string, error) {
	if a.logger != nil {
		a.logger.Info(fmt.Sprintf("Processing file: %s", inputPath))
	}

	if inputPath == "" {
		return "", fmt.Errorf("no input file selected")
	}

	dir := filepath.Dir(inputPath)
	ext := filepath.Ext(inputPath)
	base := filepath.Base(inputPath)
	name := base[:len(base)-len(ext)]

	// Determine output path based on overwrite mode
	var outputPath string
	if overwriteMode {
		// Overwrite the original file
		outputPath = inputPath
	} else {
		// Use suffix for new file name
		suffix := suffixOverride
		if suffix == "" && a.config != nil {
			suffix = a.config.Current.FileSuffix
		}
		if suffix == "" {
			suffix = "_frozen"
		}
		outputPath = filepath.Join(dir, name+suffix+".pdf")
	}

	// Use provided settings (from UI) or fallback to config/default
	prefix := prefixOverride
	if prefix == "" && a.config != nil {
		prefix = a.config.Current.Prefix
	}
	if prefix == "" {
		prefix = "AR"
	}

	position := positionOverride
	if position == "" && a.config != nil {
		position = a.config.Current.OverlayPosition
	}
	if position == "" {
		position = "bottom-right"
	}

	// Use provided compression or fallback to config
	compression := compressionLevel
	if compression == "" && a.config != nil {
		compression = a.config.Current.CompressionLevel
	}
	if compression == "" {
		compression = "none"
	}

	opts := engine.ProcessOptions{
		InputPath:        inputPath,
		OutputPath:       outputPath,
		Overlay:          overlayOverride,
		Prefix:           prefix,
		Position:         position,
		CompressionLevel: compression,
	}

	// Use background context for pipeline
	ctx := context.TODO()

	err := a.pipeline.Process(ctx, opts)
	if err != nil {
		if a.logger != nil {
			a.logger.Error(fmt.Sprintf("Process failed: %v", err))
		}
		return "", err
	}

	if a.logger != nil {
		a.logger.Info(fmt.Sprintf("Success: %s", outputPath))
	}
	return outputPath, nil
}

// GetCurrentNumber returns the next number
func (a *App) GetCurrentNumber() (int, error) {
	if a.counter == nil {
		return 0, fmt.Errorf("counter not initialized")
	}
	val, err := a.counter.GetCurrent()
	if err != nil {
		return 0, err
	}
	return val + 1, nil
}

// SetNumberOverride sets the counter config
func (a *App) SetNumberOverride(val int) error {
	if a.counter == nil {
		return fmt.Errorf("counter not initialized")
	}
	if val < 1 {
		return fmt.Errorf("number must be >= 1")
	}
	if a.logger != nil {
		a.logger.Info(fmt.Sprintf("Counter override to: %d", val))
	}
	return a.counter.SetOverride(val - 1)
}

// GetConfig returns current config
func (a *App) GetConfig() config.AppConfig {
	if a.config == nil {
		return config.DefaultConfig()
	}
	return a.config.Current
}

// SetPrefix updates the serial number prefix
func (a *App) SetPrefix(prefix string) error {
	if a.config == nil {
		return fmt.Errorf("config not initialized")
	}
	if a.logger != nil {
		a.logger.Info(fmt.Sprintf("Prefix updated to: %s", prefix))
	}
	return a.config.UpdatePrefix(prefix)
}

// SetOverlayPosition updates the serial number position
func (a *App) SetOverlayPosition(pos string) error {
	if a.config == nil {
		return fmt.Errorf("config not initialized")
	}
	if a.logger != nil {
		a.logger.Info(fmt.Sprintf("Position updated to: %s", pos))
	}
	return a.config.UpdateOverlayPosition(pos)
}

// SetCompressionLevel updates the compression level setting
func (a *App) SetCompressionLevel(level string) error {
	if a.config == nil {
		return fmt.Errorf("config not initialized")
	}
	if a.logger != nil {
		a.logger.Info(fmt.Sprintf("Compression level updated to: %s", level))
	}
	return a.config.UpdateCompressionLevel(level)
}
