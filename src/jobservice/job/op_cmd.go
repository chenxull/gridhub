package job

const (
	// StopCommand is const for stop command
	StopCommand OPCommand = "stop"
	// NilCommand is const for a nil command
	NilCommand OPCommand = "nil"
)

// OPCommand is the type of job operation commands
type OPCommand string

// IsStop return if the op command is stop
func (oc OPCommand) IsStop() bool {
	return oc == "stop"
}
