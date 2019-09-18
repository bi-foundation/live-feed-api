package models

import "fmt"

// RunState to define the state of the service
type RunState int

// Different run states
const (
	New RunState = iota
	Booting
	Running
	Stopping
	Stopped
)

// String export the run state to a string
func (runState RunState) String() string {
	switch runState {
	case New:
		return "New"
	case Booting:
		return "Booting"
	case Running:
		return "Running"
	case Stopping:
		return "Stopping"
	case Stopped:
		return "Stopped"
	default:
		return fmt.Sprintf("Unknown state %d", int(runState))
	}
}
