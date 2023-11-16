package requester

// Logger is the interface that a logger must implement
type Logger interface {
	Printf(string, ...any)
}

// nullLogger implements the logger interface but does nothing
type nullLogger struct{}

func (nullLogger) Printf(string, ...any) {
	// do nothing
}
