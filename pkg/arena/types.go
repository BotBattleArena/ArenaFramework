package arena

import "encoding/json"

// PlayerStatus represents the connection status of an input process.
type PlayerStatus int

const (
	// StatusConnected indicates the input process is running and responsive.
	StatusConnected PlayerStatus = iota
	// StatusDisconnected indicates the input process has exited.
	StatusDisconnected
	// StatusTimedOut indicates the input process did not respond in time.
	StatusTimedOut
)

// String returns a human-readable status string.
func (s PlayerStatus) String() string {
	switch s {
	case StatusConnected:
		return "connected"
	case StatusDisconnected:
		return "disconnected"
	case StatusTimedOut:
		return "timed_out"
	default:
		return "unknown"
	}
}

// Player represents a connected input process.
type Player struct {
	// ID is derived from the executable filename (without extension).
	ID string
	// Status is the current connection status.
	Status PlayerStatus
}

// Axis defines a single input axis.
type Axis struct {
	// Name is the axis identifier, e.g. "move_x", "shoot".
	Name string `json:"name"`
	// Value is the default value (typically 0.0).
	Value float32 `json:"value"`
}

// ServerMessage is sent from the game (host) to an input process (bot).
type ServerMessage struct {
	// Type of message: "start", "state", "end"
	Type string `json:"type"`

	// Game-specific state data. Using json.RawMessage allows
	// passing through pre-marshaled JSON without double encoding.
	State json.RawMessage `json:"state,omitempty"`

	// Axis definitions (sent only with type "start")
	Axes []Axis `json:"axes,omitempty"`
}

// InputMessage is sent from an input process (bot) back to the game.
type InputMessage struct {
	// Axis values set by the input, e.g. {"move_x": 0.5, "shoot": 1.0}
	Axes map[string]float32 `json:"axes"`
}

// ConnectHandler is called when an input process connects.
type ConnectHandler func(player Player)

// DisconnectHandler is called when an input process disconnects.
type DisconnectHandler func(player Player, err error)
