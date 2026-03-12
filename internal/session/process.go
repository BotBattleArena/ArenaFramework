package session

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"sync"

	pb "github.com/BotBattleArena/ArenaFramework/gen/go/arena/v1"
	"github.com/BotBattleArena/ArenaFramework/internal/protocol"
)

// Process represents a single running input executable.
type Process struct {
	ID   string
	Path string

	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser

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

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("start process %s (%s): %w", p.ID, p.Path, err)
	}

	return nil
}

// SendServerMessage writes a ServerMessage to the process's stdin.
func (p *Process) SendServerMessage(msg *pb.ServerMessage) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.stdin == nil {
		return fmt.Errorf("process %s: stdin not available", p.ID)
	}

	w := bufio.NewWriter(p.stdin)
	if err := protocol.WriteServerMessage(w, msg); err != nil {
		return fmt.Errorf("send to %s: %w", p.ID, err)
	}
	return w.Flush()
}

// ReadInputMessage reads an InputMessage from the process's stdout.
// This blocks until a message is available or the pipe is closed.
func (p *Process) ReadInputMessage() (*pb.InputMessage, error) {
	if p.stdout == nil {
		return nil, fmt.Errorf("process %s: stdout not available", p.ID)
	}
	msg, err := protocol.ReadInputMessage(p.stdout)
	if err != nil {
		return nil, fmt.Errorf("read from %s: %w", p.ID, err)
	}
	return msg, nil
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
