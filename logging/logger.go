package logging

// Logger is an interface that allows callers of the library to
// provide the library packages with logging capabilities.
//
type Logger interface {
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}
