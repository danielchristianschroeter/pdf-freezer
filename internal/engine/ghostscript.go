package engine

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// GhostscriptWrapper handles interactions with the gs CLI
type GhostscriptWrapper struct {
	ExecutablePath string // Usually "gs" or "gswin64c.exe"
}

// NewGhostscriptWrapper creates a new wrapper, detecting the executable
func NewGhostscriptWrapper() *GhostscriptWrapper {
	// Common paths where Ghostscript might be installed
	// GUI apps on macOS may not have the full shell PATH
	commonPaths := []string{
		"/usr/local/bin/gs",
		"/opt/homebrew/bin/gs",
		"/usr/bin/gs",
		"gs", // Fallback to PATH lookup
		"gswin64c.exe",
		"gswin32c.exe",
	}

	found := ""
	for _, p := range commonPaths {
		// If it's an absolute path, check if it exists
		if filepath.IsAbs(p) {
			if _, err := os.Stat(p); err == nil {
				found = p
				break
			}
		} else {
			// Use LookPath for relative names
			if path, err := exec.LookPath(p); err == nil {
				found = path
				break
			}
		}
	}

	if found == "" {
		found = "gs" // Last resort fallback
	}

	return &GhostscriptWrapper{ExecutablePath: found}
}

// CheckDependencies verifies if Ghostscript is installed and runnable
func (g *GhostscriptWrapper) CheckDependencies() error {
	cmd := exec.Command(g.ExecutablePath, "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ghostscript not found or not working: %w", err)
	}
	return nil
}

// ExtractPages extracts all pages from pdfPath as JPEG images into outDir.
// dpi controls resolution (lower = smaller files), quality controls JPEG compression (1-100).
// Returns the list of generated image files sorted by page number.
func (g *GhostscriptWrapper) ExtractPages(ctx context.Context, pdfPath, outDir string, dpi, quality int) ([]string, error) {
	// Ensure absolute paths
	absPdf, err := filepath.Abs(pdfPath)
	if err != nil {
		return nil, err
	}
	absOut, err := filepath.Abs(outDir)
	if err != nil {
		return nil, err
	}

	// Output pattern: page-%d.jpg
	// %d will be replaced by 1-based page number by GS
	outPattern := filepath.Join(absOut, "page-%d.jpg")

	// Construct command
	// gs -dNOPAUSE -dBATCH -sDEVICE=jpeg -dJPEGQ=quality -rDPI -sOutputFile=... input.pdf
	args := []string{
		"-dNOPAUSE",
		"-dBATCH",
		"-dSAFER",
		"-sDEVICE=jpeg",
		fmt.Sprintf("-dJPEGQ=%d", quality),
		fmt.Sprintf("-r%d", dpi),
		fmt.Sprintf("-sOutputFile=%s", outPattern),
		absPdf,
	}

	cmd := exec.CommandContext(ctx, g.ExecutablePath, args...)
	// Capture stderr for debugging if needed, but for now just run
	// GS writes info to stdout/stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ghostscript failed: %w\nOutput: %s", err, string(output))
	}

	// Identify generated files
	// Since GS uses %d, filenames will be page-1.jpg, page-2.jpg, etc.
	// We read the directory to find them.
	files, err := os.ReadDir(absOut)
	if err != nil {
		return nil, err
	}

	var savedFiles []string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "page-") && strings.HasSuffix(f.Name(), ".jpg") {
			savedFiles = append(savedFiles, filepath.Join(absOut, f.Name()))
		}
	}

	// Sort files to ensure page order (page-1.jpg, page-2.jpg, page-10.jpg)
	// Need custom sort because string sort puts page-10 before page-2
	// For simpler logic, rely on standard sort if we verify zero-padding? No GS output %d is not zero padded by default.
	// We should sort by number.
	// Simple fix: use a number-aware sort or parse numbers.
	// Or use %04d in pattern?
	// Let's use %04d to simplify sorting. wrapper handles it.

	// Refine pattern:
	// actually GS supports printf syntax.
	// If the user didn't specify, we use %d.
	// We can use page-%04d.jpg? GS supports standard printf. Let's try.
	// If not supported, we implement custom sort.
	// Standard `gs` supports `%d`. Some versions support `%04d`.
	// For maximum compatibility, assume `%d` and sort manually.

	sort.Slice(savedFiles, func(i, j int) bool {
		// Extract numbers logic... or just use natural sort library.
		// Since we don't want deps, we'll try to rely on file creation time? Unreliable.
		// We'll parse the filename.
		return extractPageNum(savedFiles[i]) < extractPageNum(savedFiles[j])
	})

	return savedFiles, nil
}

func extractPageNum(path string) int {
	base := filepath.Base(path)
	// format: page-X.jpg
	var num int
	fmt.Sscanf(base, "page-%d.jpg", &num)
	return num
}
