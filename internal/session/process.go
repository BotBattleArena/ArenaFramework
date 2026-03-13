package session

import (
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/BotBattleArena/ArenaFramework/internal/protocol"
)

// Process represents a single running input executable.
type Process struct {
	ID   string
	Path string

	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser

	encoder *protocol.Encoder
	decoder *protocol.Decoder

	mu sync.Mutex
}

// NewProcess creates a new Process for the given executable path.
func NewProcess(id, path string) *Process {
	return &Process{
		ID:   id,
		Path: path,
	}
}

// Start launches the executable as a subprocess and sets up stdin/stdout pipes.
func (p *Process) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.cmd = exec.Command(p.Path)

	var err error
	p.stdin, err = p.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("create stdin pipe for %s: %w", p.ID, err)
	}

	p.stdout, err = p.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create stdout pipe for %s: %w", p.ID, err)
	}

	p.encoder = protocol.NewEncoder(p.stdin)
	p.decoder = protocol.NewDecoder(p.stdout)

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("start process %s (%s): %w", p.ID, p.Path, err)
	}

	return nil
}

// SendMessage writes a JSON message to the process's stdin.
func (p *Process) SendMessage(msg interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.encoder == nil {
		return fmt.Errorf("process %s: encoder not available", p.ID)
	}

	if err := p.encoder.Encode(msg); err != nil {
		return fmt.Errorf("send to %s: %w", p.ID, err)
	}
	return nil
}

// ReadMessage reads a JSON message from the process's stdout into the given value.
// This blocks until a message is available or the pipe is closed.
func (p *Process) ReadMessage(v interface{}) error {
	if p.decoder == nil {
		return fmt.Errorf("process %s: decoder not available", p.ID)
	}
	if err := p.decoder.Decode(v); err != nil {
		return fmt.Errorf("read from %s: %w", p.ID, err)
	}
	return nil
}

// Stop terminates the subprocess gracefully.
func (p *Process) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.stdin != nil {
		p.stdin.Close()
	}
	if p.cmd != nil && p.cmd.Process != nil {
		return p.cmd.Process.Kill()
	}
	return nil
}

// Wait waits for the subprocess to exit and returns the error (if any).
func (p *Process) Wait() error {
	if p.cmd == nil {
		return nil
	}
	return p.cmd.Wait()
}
