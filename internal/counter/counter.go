package counter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// CounterState represents the persisted state
type CounterState struct {
	Current int `json:"current"`
}

// Manager handles the persistent counter
type Manager struct {
	mu         sync.Mutex
	configPath string
	lockPath   string
	statePath  string
	locked     bool
}

// NewManager creates a new counter manager
func NewManager() (*Manager, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user config dir: %w", err)
	}

	appDir := filepath.Join(configDir, "pdf-freezer")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config dir: %w", err)
	}

	return &Manager{
		configPath: appDir,
		statePath:  filepath.Join(appDir, "counter.json"),
		lockPath:   filepath.Join(appDir, "counter.lock"),
	}, nil
}

// Lock attempts to acquire an exclusive lock for batch processing
// It creates a .lock file. If it exists, it returns an error.
func (m *Manager) Lock() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.locked {
		return fmt.Errorf("already locked by this instance")
	}

	// Try to create lock file specifically
	f, err := os.OpenFile(m.lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("counter is locked by another instance or process")
		}
		return err
	}
	f.Close()

	m.locked = true
	return nil
}

// Unlock releases the lock
func (m *Manager) Unlock() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.locked {
		return nil
	}

	if err := os.Remove(m.lockPath); err != nil {
		return fmt.Errorf("failed to remove lock file: %w", err)
	}

	m.locked = false
	return nil
}

// GetCurrent returns the current counter value without incrementing
func (m *Manager) GetCurrent() (int, error) {
	state, err := m.loadState()
	if err != nil {
		return 0, err
	}
	return state.Current, nil
}

// GetNext increments the counter and returns the NEW value
// It automatically persists the change.
func (m *Manager) GetNext() (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. Load
	state, err := m.loadState()
	if err != nil {
		return 0, err
	}

	// 2. Increment
	state.Current++

	// 3. Save
	if err := m.saveState(state); err != nil {
		return 0, err
	}

	return state.Current, nil
}

// SetOverride forces the counter to a specific value
func (m *Manager) SetOverride(val int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state := CounterState{Current: val}
	return m.saveState(state)
}

func (m *Manager) loadState() (CounterState, error) {
	data, err := os.ReadFile(m.statePath)
	if os.IsNotExist(err) {
		// Default start at 0 (first GetNext will be 1)
		return CounterState{Current: 0}, nil
	}
	if err != nil {
		return CounterState{}, err
	}

	var state CounterState
	if err := json.Unmarshal(data, &state); err != nil {
		return CounterState{}, err
	}
	return state, nil
}

func (m *Manager) saveState(state CounterState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.statePath, data, 0644)
}

// ForceUnlock cleans up a stale lock file (use with caution, maybe on startup if requested)
func (m *Manager) ForceUnlock() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if file status indicates it's very old?
	// For now, just remove it as requested.
	if err := os.Remove(m.lockPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	m.locked = false
	return nil
}
