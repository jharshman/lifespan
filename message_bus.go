package lifespan

import "time"

// Error is a message that can be sent via a lifespan MessageBus.
// This type contains a standard error with additional metadata like the lifespan's UUID.
type Error struct {
	JobID     string
	GroupID   string
	Error     error
	Timestamp time.Time
}

// Log is a message that can be sent via a lifespan MessageBus.
// This type contains a standard error with additional metadata like the lifespan's UUID.
type Log struct {
	JobID     string
	GroupID   string
	Msg       string
	Level     string
	Metadata  map[string]any
	Timestamp time.Time
}

// MessageBus defines behavior for a generic message bus.
// The implementations within the lifespan package provide
// a MessageBus for Errors and a MessageBus for Logs.
type MessageBus[T any] interface {
	Publish(msg T)
	Subscribe() <-chan T
	Close()
}

// ErrorBus implements the MessageBus interface for the Error type.
type ErrorBus struct {
	bus chan Error
}

// NewErrorBus returns a pointer to ErrorBus with the given buffer size.
// If bsize is less than defaultBufferSize, set bsize to defaultBufferSize.
func NewErrorBus(bsize int64) *ErrorBus {
	if bsize < defaultBufferSize {
		bsize = defaultBufferSize
	}
	return &ErrorBus{
		bus: make(chan Error, bsize),
	}
}

// Publish writes the Error to the bus.
func (e *ErrorBus) Publish(msg Error) {
	select {
	case e.bus <- msg:
	default:
		// todo: record dropped messages
	}
}

// Subscribe returns a receive channel for the ErrorBus implementation.
func (e *ErrorBus) Subscribe() <-chan Error {
	return e.bus
}

// Close will close the channel contained within ErrorBus.
func (e *ErrorBus) Close() {
	close(e.bus)
}

// LogBus implements the MessageBus interface for the Log type.
type LogBus struct {
	bus chan Log
}

// NewLogBus returns a pointer to *LogBus with the given buffer size.
// If bsize is less than defaultBufferSize, set bsize to defaultBufferSize.
func NewLogBus(bsize int64) *LogBus {
	if bsize < defaultBufferSize {
		bsize = defaultBufferSize
	}
	return &LogBus{
		bus: make(chan Log, bsize),
	}
}

// Publish writes the Log to the bus.
func (l *LogBus) Publish(msg Log) {
	select {
	case l.bus <- msg:
	default:
		// todo: record dropped messages
	}
}

// Subscribe returns a receive channel for the LogBus implementation.
func (l *LogBus) Subscribe() <-chan Log {
	return l.bus
}

// Close will close the channel contained within LogBus.
func (l *LogBus) Close() {
	close(l.bus)
}
