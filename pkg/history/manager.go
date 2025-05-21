package history

import (
"encoding/json"
"fmt"
"os"
"path/filepath"
)

// Manager handles request history operations
type Manager struct {
historyPath string
}

// NewManager creates a new history manager
func NewManager(baseDir string) (*Manager, error) {
historyPath := filepath.Join(baseDir, "history.json")
return &Manager{
historyPath: historyPath,
}, nil
}

// LoadHistory loads the request history from file
func (m *Manager) LoadHistory() (*History, error) {
data, err := os.ReadFile(m.historyPath)
if err != nil {
if os.IsNotExist(err) {
return &History{}, nil
}
return nil, fmt.Errorf("reading history file: %w", err)
}

var history History
if err := json.Unmarshal(data, &history); err != nil {
return nil, fmt.Errorf("parsing history file: %w", err)
}

return &history, nil
}

// SaveHistory saves the request history to file
func (m *Manager) SaveHistory(history *History) error {
data, err := json.MarshalIndent(history, "", "  ")
if err != nil {
return fmt.Errorf("serializing history: %w", err)
}

if err := os.WriteFile(m.historyPath, data, 0644); err != nil {
return fmt.Errorf("writing history file: %w", err)
}

return nil
}

// AddEntry adds a new entry to the history
func (m *Manager) AddEntry(entry Entry) error {
history, err := m.LoadHistory()
if err != nil {
return err
}

history.Entries = append([]Entry{entry}, history.Entries...)

// Keep only the last 100 entries
if len(history.Entries) > 100 {
history.Entries = history.Entries[:100]
}

return m.SaveHistory(history)
}

// GetEntry gets a specific history entry by ID
func (m *Manager) GetEntry(id string) (*Entry, error) {
history, err := m.LoadHistory()
if err != nil {
return nil, err
}

for _, entry := range history.Entries {
if entry.ID == id {
return &entry, nil
}
}

return nil, fmt.Errorf("entry not found: %s", id)
}

// ListEntries lists history entries with optional limit
func (m *Manager) ListEntries(limit int) ([]Entry, error) {
history, err := m.LoadHistory()
if err != nil {
return nil, err
}

if limit > 0 && limit < len(history.Entries) {
return history.Entries[:limit], nil
}

return history.Entries, nil
}

// ClearHistory clears all history entries
func (m *Manager) ClearHistory() error {
return m.SaveHistory(&History{})
}
