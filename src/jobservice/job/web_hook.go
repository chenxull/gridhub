package job

// HookCallback defines a callback to trigger when hook events happened
type HookCallback func(hookURL string, change *StatusChange) error
