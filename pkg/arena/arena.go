package arena

import (
	"errors"
	"fmt"
	"sync"
	"time"

	pb "github.com/BotBattleArena/ArenaFramework/gen/go/arena/v1"
	"github.com/BotBattleArena/ArenaFramework/internal/session"
)

// Sentinel errors.
var (
	ErrNoInputDir     = errors.New("arena: InputDir must be set")
	ErrNoAxes         = errors.New("arena: at least one axis must be defined")
	ErrNotRunning     = errors.New("arena: not running")
	ErrAlreadyRunning = errors.New("arena: already running")
)

// Arena is the main framework instance that manages input processes.
type Arena struct {
	cfg     Config
	manager *session.Manager
	running bool

	onAxes       AxesHandler
	onConnect    ConnectHandler
	onDisconnect DisconnectHandler

	mu sync.RWMutex
}

// New creates a new Arena instance with the given configuration.
func New(cfg Config) (*Arena, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &Arena{
		cfg:     cfg,
		manager: session.NewManager(),
	}, nil
}

// Start loads all input executables from InputDir, starts them as subprocesses,
// and sends the initial "start" message with axis definitions.
func (a *Arena) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.running {
		return ErrAlreadyRunning
	}

	// Load executables from input directory
	if err := a.manager.LoadFromDir(a.cfg.InputDir); err != nil {
		return fmt.Errorf("load inputs: %w", err)
	}

	// Start all processes
	if err := a.manager.StartAll(); err != nil {
		a.manager.StopAll()
		return fmt.Errorf("start inputs: %w", err)
	}

	a.running = true

	// Notify connect handlers
	if a.onConnect != nil {
		for _, id := range a.manager.ProcessIDs() {
			a.onConnect(Player{ID: id, Status: StatusConnected})
		}
	}

	// Send start message with axis definitions
	startMsg := a.buildStartMessage()
	if err := a.manager.SendToAll(startMsg); err != nil {
		return fmt.Errorf("send start message: %w", err)
	}

	return nil
}

// Stop terminates all input processes.
func (a *Arena) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		return
	}

	// Send end message (best effort)
	endMsg := &pb.ServerMessage{Type: "end"}
	a.manager.SendToAll(endMsg)

	a.manager.StopAll()
	a.running = false
}

// SendState sends game state data to all input processes.
func (a *Arena) SendState(state []byte) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.running {
		return ErrNotRunning
	}

	msg := &pb.ServerMessage{
		Type:  "state",
		State: state,
	}
	return a.manager.SendToAll(msg)
}

// SendStateTo sends game state data to a specific input process.
func (a *Arena) SendStateTo(inputID string, state []byte) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.running {
		return ErrNotRunning
	}

	msg := &pb.ServerMessage{
		Type:  "state",
		State: state,
	}
	return a.manager.SendTo(inputID, msg)
}

// RequestAxes sends the current state to all inputs and waits for their axis
// responses within the given timeout. Returns a map of inputID -> axis values.
// Inputs that don't respond in time get default axis values.
func (a *Arena) RequestAxes(state []byte, timeout time.Duration) map[string]map[string]float32 {
	if timeout == 0 {
		timeout = a.cfg.ActionTimeout
	}

	ids := a.manager.ProcessIDs()
	result := make(map[string]map[string]float32, len(ids))
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Send state to all
	msg := &pb.ServerMessage{
		Type:  "state",
		State: state,
	}
	a.manager.SendToAll(msg)

	// Wait for responses in parallel
	for _, id := range ids {
		wg.Add(1)
		go func(inputID string) {
			defer wg.Done()

			ch := make(chan map[string]float32, 1)
			go func() {
				resp, err := a.manager.ReadFrom(inputID)
				if err != nil {
					ch <- a.defaultAxes()
					return
				}
				ch <- resp.GetAxes()
			}()

			select {
			case axes := <-ch:
				mu.Lock()
				result[inputID] = axes
				mu.Unlock()
			case <-time.After(timeout):
				mu.Lock()
				result[inputID] = a.defaultAxes()
				mu.Unlock()
			}
		}(id)
	}

	wg.Wait()
	return result
}

// Players returns a list of all registered input processes.
func (a *Arena) Players() []Player {
	ids := a.manager.ProcessIDs()
	players := make([]Player, len(ids))
	for i, id := range ids {
		players[i] = Player{
			ID:     id,
			Status: StatusConnected, // TODO: track actual status
		}
	}
	return players
}

// Player returns a single input by ID.
func (a *Arena) Player(id string) (Player, bool) {
	_, ok := a.manager.GetProcess(id)
	if !ok {
		return Player{}, false
	}
	return Player{ID: id, Status: StatusConnected}, true
}

// Axes returns the configured axes.
func (a *Arena) Axes() []Axis {
	return a.cfg.Axes
}

// IsRunning returns whether the arena is currently active.
func (a *Arena) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}

// OnAxes registers a handler for incoming axis values from inputs.
func (a *Arena) OnAxes(handler AxesHandler) {
	a.onAxes = handler
}

// OnConnect registers a handler for when an input process connects.
func (a *Arena) OnConnect(handler ConnectHandler) {
	a.onConnect = handler
}

// OnDisconnect registers a handler for when an input process disconnects.
func (a *Arena) OnDisconnect(handler DisconnectHandler) {
	a.onDisconnect = handler
}

// --- internal helpers ---

func (a *Arena) buildStartMessage() *pb.ServerMessage {
	axes := make([]*pb.Axis, len(a.cfg.Axes))
	for i, ax := range a.cfg.Axes {
		axes[i] = &pb.Axis{
			Name:  ax.Name,
			Value: ax.Value,
		}
	}
	return &pb.ServerMessage{
		Type: "start",
		Axes: axes,
	}
}

func (a *Arena) defaultAxes() map[string]float32 {
	defaults := make(map[string]float32, len(a.cfg.Axes))
	for _, ax := range a.cfg.Axes {
		defaults[ax.Name] = ax.Value
	}
	return defaults
}

func (a *Arena) readLoop(inputID string) {
	for {
		resp, err := a.manager.ReadFrom(inputID)
		if err != nil {
			// Process likely exited
			if a.onDisconnect != nil {
				a.onDisconnect(Player{ID: inputID, Status: StatusDisconnected}, err)
			}
			return
		}

		if a.onAxes != nil {
			a.onAxes(Player{ID: inputID, Status: StatusConnected}, resp.GetAxes())
		}
	}
}
