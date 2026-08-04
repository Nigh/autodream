package action

// Kind identifies an executable action family.
type Kind string

const (
	KindMouseMove   Kind = "mouse_move"
	KindMouseClick  Kind = "mouse_click"
	KindMouseButton Kind = "mouse_button"
	KindKey         Kind = "key"
	KindKeyTap      Kind = "key_tap"
	KindGamepad     Kind = "gamepad"
	KindNoop        Kind = "noop"
)

// Action is a platform-agnostic command for Executor implementations.
type Action struct {
	Kind Kind

	// Mouse
	X, Y   int
	Button string // left, right, middle
	Down   bool   // for mouse_button / key hold semantics

	// Keyboard
	Key      string
	Modifiers []string

	// Gamepad
	GamepadIndex int
	Control      string
	Value        float64

	// Opaque extension for custom executors
	Extra map[string]any
}
