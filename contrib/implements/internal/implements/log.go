package implements

// DebugLogger is a simple interface to allow debug logging w/in the
// implements package.
type DebugLogger interface {
	Printf(format string, v ...interface{})
}

// NoOpLogger is a dummy logger that generates no output.
type NoOpLogger struct{}

// Printf implements the DebugLogger interface.
func (NoOpLogger) Printf(_ string, _ ...interface{}) {}

var logger DebugLogger = NoOpLogger{}

// SetLogger will set the given debug logger for the whole implements package.
func SetLogger(l DebugLogger) {
	logger = l
}
