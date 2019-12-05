package logging

// StubLogger is a type that meets the Logger interface but goes nowhere and
// does nothing.
type StubLogger struct{}

func NewStubLogger() StubLogger {
	return StubLogger{}
}

func (StubLogger) Errorf(format string, args ...interface{}) {
	return
}

func (StubLogger) Infof(format string, args ...interface{}) {
	return
}

func (StubLogger) Debugf(format string, args ...interface{}) {
	return
}
