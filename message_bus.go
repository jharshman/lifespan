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
	Bus chan Error
}

// Publish writes the Error to the Bus.
func (e *ErrorBus) Publish(msg Error) {
	select {
	case e.Bus <- msg:
	default:
	}
}

// Subscribe returns a receive channel for the ErrorBus implementation.
func (e *ErrorBus) Subscribe() <-chan Error {
	return e.Bus
}

// Close will close the channel contained within ErrorBus.
func (e *ErrorBus) Close() {
	close(e.Bus)
}

// LogBus implements the MessageBus interface for the Log type.
type LogBus struct {
	Bus chan Log
}

// Publish writes the Log to the Bus.
func (l *LogBus) Publish(msg Log) {
	select {
	case l.Bus <- msg:
	default:
	}
}

// Subscribe returns a receive channel for the LogBus implementation.
func (l *LogBus) Subscribe() <-chan Log {
	return l.Bus
}

// Close will close the channel contained within LogBus.
func (l *LogBus) Close() {
	close(l.Bus)
}
