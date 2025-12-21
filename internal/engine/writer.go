package engine

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"os"

	"github.com/signintech/gopdf"
)

// PDFWriter reconstructs the PDF
type PDFWriter struct {
	pdf *gopdf.GoPdf
}

// NewPDFWriter creates a new writer instance
func NewPDFWriter(fontPath string) *PDFWriter {
	pdf := &gopdf.GoPdf{}

	// Start document (A4 default, overridden per page)
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	// Load Font
	// We will use the font provided.
	if fontPath != "" {
		err := pdf.AddTTFFont("Inter", fontPath)
		if err != nil {
			fmt.Printf("Warning: Failed to load font: %v\n", err)
		}
	}

	return &PDFWriter{pdf: pdf}
}

// AddPage adds a JPEG image as a page.
// If overlayText is not empty, it prints it at a configured position.
func (w *PDFWriter) AddPage(imagePath string, overlayText string, position string, dpi int) error {
	// ... decoding config ...
	f, err := os.Open(imagePath)
	if err != nil {
		return err
	}
	defer f.Close()

	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return err
	}

	widthPt := float64(cfg.Width) * 72.0 / float64(dpi)
	heightPt := float64(cfg.Height) * 72.0 / float64(dpi)

	w.pdf.AddPageWithOption(gopdf.PageOption{
		PageSize: &gopdf.Rect{W: widthPt, H: heightPt},
	})

	err = w.pdf.Image(imagePath, 0, 0, &gopdf.Rect{W: widthPt, H: heightPt})
	if err != nil {
		return err
	}

	if overlayText != "" {
		if err := w.pdf.SetFont("Inter", "", 12); err != nil {
			return fmt.Errorf("failed to set font: %w", err)
		}
		w.pdf.SetTextColor(255, 0, 0)

		textWidth, err := w.pdf.MeasureTextWidth(overlayText)
		if err != nil {
			return err
		}

		margin := 1.0
		var x, y float64

		switch position {
		case "top-left":
			x = margin
			y = margin + 12 // Font height offset approximately
		case "top-right":
			x = widthPt - textWidth - margin
			y = margin + 12
		case "bottom-left":
			x = margin
			y = heightPt - margin
		case "bottom-right":
			fallthrough
		default:
			x = widthPt - textWidth - margin
			y = heightPt - margin
		}

		w.pdf.SetX(x)
		w.pdf.SetY(y)

		if err := w.pdf.Cell(nil, overlayText); err != nil {
			return err
		}
	}
	return nil
}

// Save writes the PDF to disk
func (w *PDFWriter) Save(path string) error {
	return w.pdf.WritePdf(path)
}
