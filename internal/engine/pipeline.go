package engine

import (
	"context"
	"fmt"
	"os"
	"sync"

	"pdf-freezer/internal/config"
	"pdf-freezer/internal/counter"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate buffer for standard page size ~2MB?
		return make([]byte, 0, 2*1024*1024)
	},
}

// Pipeline Orchestrates the freezing process
type Pipeline struct {
	counter *counter.Manager
	gs      *GhostscriptWrapper
}

// NewPipeline creates a new pipeline
func NewPipeline(c *counter.Manager) *Pipeline {
	return &Pipeline{
		counter: c,
		gs:      NewGhostscriptWrapper(),
	}
}

// ProcessOptions configuration for a job
type ProcessOptions struct {
	InputPath        string
	OutputPath       string
	Overlay          bool
	Prefix           string
	Position         string
	CompressionLevel string // none, low, medium, high
}

// Process executes the freeze pipeline
func (p *Pipeline) Process(ctx context.Context, opts ProcessOptions) error {
	// 1. Check dependencies
	if err := p.gs.CheckDependencies(); err != nil {
		return err
	}

	// 2. Lock Counter (Batch scope? Or per file? Usually per file or batch.
	// For this single process call, we lock around the number generation.

	usageNum, err := p.counter.GetNext()
	if err != nil {
		return fmt.Errorf("counter error: %w", err)
	}

	// 3. Create Temp Dir for pages
	tmpDir, err := os.MkdirTemp("", "pdf-freezer-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	// 4. Extract Pages with compression settings
	// Pass context for cancellation/timeout
	compSettings := config.GetCompressionSettings(opts.CompressionLevel)
	images, err := p.gs.ExtractPages(ctx, opts.InputPath, tmpDir, compSettings.DPI, compSettings.Quality)
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	if len(images) == 0 {
		return fmt.Errorf("no pages extracted")
	}

	// 5. Initialize Writer
	// We need to write the embedded font to a temp file because gopdf requires a path
	fontTmp, err := os.CreateTemp("", "font-*.ttf")
	if err != nil {
		return fmt.Errorf("failed to create font temp file: %w", err)
	}
	defer os.Remove(fontTmp.Name()) // Clean up

	if _, err := fontTmp.Write(InterFontData); err != nil {
		return fmt.Errorf("failed to write font data: %w", err)
	}
	fontTmp.Close()

	writer := NewPDFWriter(fontTmp.Name())

	// 6. Re-assemble
	prefix := opts.Prefix
	if prefix == "" {
		prefix = "AR" // Default fallback
	}
	serialText := fmt.Sprintf("%s%04d", prefix, usageNum)

	for i, imgPath := range images {
		// Overlay only on first page
		txt := ""
		if i == 0 && opts.Overlay {
			txt = serialText
		}

		if err := writer.AddPage(imgPath, txt, opts.Position, compSettings.DPI); err != nil {
			return fmt.Errorf("failed to write page %d: %w", i+1, err)
		}
	}

	// 7. Save
	if err := writer.Save(opts.OutputPath); err != nil {
		return fmt.Errorf("failed to save output: %w", err)
	}

	return nil
}
