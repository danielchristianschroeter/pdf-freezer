package app

import (
	"testing"
)

func TestNewApp(t *testing.T) {
	app := NewApp()
	if app == nil {
		t.Fatal("NewApp returned nil")
	}
	if app.counter == nil {
		t.Error("App counter not initialized")
	}
	if app.config == nil {
		t.Error("App config not initialized")
	}
}

func TestOnFileDrop(t *testing.T) {
	app := NewApp()
	// Should not panic even if application.Get() returns nil
	paths := []string{"/tmp/test.pdf"}
	app.OnFileDrop(paths)
}

func TestCounterOverride(t *testing.T) {
	app := NewApp()
	// Set override
	err := app.SetNumberOverride(100)
	if err != nil {
		t.Fatalf("SetNumberOverride failed: %v", err)
	}

	// Get next
	val, err := app.GetCurrentNumber()
	if err != nil {
		t.Fatalf("GetCurrentNumber failed: %v", err)
	}
	if val != 100 {
		t.Errorf("Expected 100, got %d", val)
	}
}

func TestPrefixConfig(t *testing.T) {
	app := NewApp()
	err := app.SetPrefix("TEST")
	if err != nil {
		t.Fatalf("SetPrefix failed: %v", err)
	}

	cfg := app.GetConfig()
	if cfg.Prefix != "TEST" {
		t.Errorf("Expected prefix TEST, got %s", cfg.Prefix)
	}
}
