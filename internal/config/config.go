package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// AppConfig holds persistent application settings
type AppConfig struct {
	Prefix           string `json:"prefix"`
	Overlay          bool   `json:"overlay"`
	OverlayColor     string `json:"overlay_color"`     // Hex or Name
	OverlayPosition  string `json:"overlay_position"`  // top-right, top-left, bottom-right, bottom-left
	FileSuffix       string `json:"file_suffix"`       // Suffix for output file, e.g. "_frozen"
	OverwriteMode    bool   `json:"overwrite_mode"`    // If true, overwrite original file
	CompressionLevel string `json:"compression_level"` // none, low, medium, high
}

// Manager handles config persistence
type Manager struct {
	mu         sync.RWMutex
	configPath string
	Current    AppConfig
}

// DefaultConfig returns safe defaults
func DefaultConfig() AppConfig {
	return AppConfig{
		Prefix:           "AR",
		Overlay:          true,
		OverlayColor:     "#FF0000",
		OverlayPosition:  "bottom-right",
		FileSuffix:       "_frozen",
		OverwriteMode:    false,
		CompressionLevel: "none",
	}
}

// CompressionSettings holds DPI and JPEG quality for a compression level
type CompressionSettings struct {
	DPI     int
	Quality int
}

// GetCompressionSettings returns DPI and quality values for a compression level
func GetCompressionSettings(level string) CompressionSettings {
	switch level {
	case "low":
		return CompressionSettings{DPI: 200, Quality: 85}
	case "medium":
		return CompressionSettings{DPI: 150, Quality: 75}
	case "high":
		return CompressionSettings{DPI: 100, Quality: 65}
	default: // "none" or invalid
		return CompressionSettings{DPI: 300, Quality: 95}
	}
}

// ... existing methods ...

// UpdateOverlayPosition updates and saves position
func (m *Manager) UpdateOverlayPosition(pos string) error {
	m.mu.Lock()
	m.Current.OverlayPosition = pos
	m.mu.Unlock()
	return m.Save()
}

// NewManager creates a new config manager
func NewManager() (*Manager, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	appDir := filepath.Join(configDir, "pdf-freezer")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return nil, err
	}

	m := &Manager{
		configPath: filepath.Join(appDir, "config.json"),
		Current:    DefaultConfig(),
	}

	if err := m.Load(); err != nil {
		// If load fails (e.g. no file), we use defaults which are set.
		// We might want to save the defaults?
		_ = m.Save()
	}

	return m, nil
}

// Load reads config from disk
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &m.Current)
}

// Save writes config to disk
func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := json.MarshalIndent(m.Current, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.configPath, data, 0644)
}

// UpdatePrefix updates and saves prefix
func (m *Manager) UpdatePrefix(prefix string) error {
	m.mu.Lock()
	m.Current.Prefix = prefix
	m.mu.Unlock()
	return m.Save()
}

// UpdateOverlay settings
func (m *Manager) UpdateOverlay(enabled bool) error {
	m.mu.Lock()
	m.Current.Overlay = enabled
	m.mu.Unlock()
	return m.Save()
}

// UpdateCompressionLevel updates and saves compression level
func (m *Manager) UpdateCompressionLevel(level string) error {
	// Validate level
	validLevels := map[string]bool{"none": true, "low": true, "medium": true, "high": true}
	if !validLevels[level] {
		level = "none"
	}
	m.mu.Lock()
	m.Current.CompressionLevel = level
	m.mu.Unlock()
	return m.Save()
}
