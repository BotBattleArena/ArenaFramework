package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	pb "github.com/BotBattleArena/ArenaFramework/gen/go/arena/v1"
)

// Manager manages all input processes (bot executables).
type Manager struct {
	processes map[string]*Process
	mu        sync.RWMutex
}

// NewManager creates a new session manager.
func NewManager() *Manager {
	return &Manager{
		processes: make(map[string]*Process),
	}
}

// LoadFromDir scans a directory for executable files and registers them.
// The input ID is derived from the filename without extension.
func (m *Manager) LoadFromDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read input dir %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := filepath.Ext(name)

		// Accept .exe files on Windows, or any executable
		if ext != ".exe" && ext != "" {
			continue
		}

		id := strings.TrimSuffix(name, ext)
		fullPath := filepath.Join(dir, name)

		m.mu.Lock()
		m.processes[id] = NewProcess(id, fullPath)
		m.mu.Unlock()
	}

	if len(m.processes) == 0 {
		return fmt.Errorf("no input executables found in %s", dir)
	}

	return nil
}

// StartAll launches all registered input processes.
func (m *Manager) StartAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for id, proc := range m.processes {
		if err := proc.Start(); err != nil {
			return fmt.Errorf("start input %s: %w", id, err)
		}
	}
	return nil
}

// StopAll terminates all running input processes.
func (m *Manager) StopAll() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, proc := range m.processes {
		proc.Stop()
	}
}

// SendToAll sends a ServerMessage to all input processes.
func (m *Manager) SendToAll(msg *pb.ServerMessage) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errs []error
	for _, proc := range m.processes {
		if err := proc.SendServerMessage(msg); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("send to all: %d errors, first: %w", len(errs), errs[0])
	}
	return nil
}

// SendTo sends a ServerMessage to a specific input process.
func (m *Manager) SendTo(id string, msg *pb.ServerMessage) error {
	m.mu.RLock()
	proc, ok := m.processes[id]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("unknown input: %s", id)
	}
	return proc.SendServerMessage(msg)
}

// ReadFrom reads an InputMessage from a specific input process.
// This blocks until a message is received.
func (m *Manager) ReadFrom(id string) (*pb.InputMessage, error) {
	m.mu.RLock()
	proc, ok := m.processes[id]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown input: %s", id)
	}
	return proc.ReadInputMessage()
}

// GetProcess returns a process by ID.
func (m *Manager) GetProcess(id string) (*Process, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	proc, ok := m.processes[id]
	return proc, ok
}

// ProcessIDs returns a list of all registered input IDs.
func (m *Manager) ProcessIDs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.processes))
	for id := range m.processes {
		ids = append(ids, id)
	}
	return ids
}
