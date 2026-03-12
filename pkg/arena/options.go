package arena

import "time"

// Config holds the configuration for an Arena instance.
type Config struct {
	// InputDir is the directory containing input executables.
	// All .exe files in this directory will be loaded as inputs.
	InputDir string

	// Axes defines the input axes available in the game.
	Axes []Axis

	// ActionTimeout is the maximum time to wait for a response from an input.
	// Defaults to 5 seconds if not set.
	ActionTimeout time.Duration
}

// validate checks the config and sets defaults.
func (c *Config) validate() error {
	if c.InputDir == "" {
		return ErrNoInputDir
	}
	if len(c.Axes) == 0 {
		return ErrNoAxes
	}
	if c.ActionTimeout == 0 {
		c.ActionTimeout = 5 * time.Second
	}
	return nil
}
