package lifespan

import (
	"context"
	"log/slog"
)

// defaultBufferSize provides a sane default for the underlying LogBus.
const defaultBufferSize = 1024

// Logger implements of log/slog Handler.
// It primarily sits on top of a LogBus and provides a standard log interface for it.
type Logger struct {
	opts  Options
	bus   *LogBus
	attrs []slog.Attr
}

// Options encapsulates options for the logger.
type Options struct {
	Level slog.Leveler
}

// NewLogger returns a pointer to a Logger. If the provided bsize is less than the defaultBufferSize, it will default to the
// defaultBufferSize.
func NewLogger(bsize int64, opts *Options) *Logger {
	if bsize < defaultBufferSize {
		bsize = defaultBufferSize
	}

	l := &Logger{
		opts: *opts,
		bus:  NewLogBus(bsize),
	}

	return l
}

// Enabled returns true if the requested level is greater than or equal to the minimum configured level for the Logger.
func (l *Logger) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= l.opts.Level.Level()
}

// Handle takes log records and sends them to the underlying LogBus.
func (l *Logger) Handle(ctx context.Context, r slog.Record) error {

	log := Log{
		Msg:      r.Message,
		Level:    r.Level.String(),
		Metadata: make(map[string]any, r.NumAttrs()),
	}

	// set timestamp in UTC
	if !r.Time.IsZero() {
		log.Timestamp = r.Time.UTC()
	}

	// Process the logger's stored attributes first (from WithAttrs calls)
	for _, attr := range l.attrs {
		v := attr.Value.Resolve()
		switch attr.Key {
		case "job_id":
			log.JobID = v.String()
		case "group_id":
			log.GroupID = v.String()
		default:
			log.Metadata[attr.Key] = v.Any()
		}
	}

	// Then extract attributes from slog.Record
	r.Attrs(func(attr slog.Attr) bool {
		v := attr.Value.Resolve()
		switch attr.Key {
		case "job_id":
			log.JobID = v.String()
		case "group_id":
			log.GroupID = v.String()
		default:
			log.Metadata[attr.Key] = v.Any()
		}
		return true
	})

	l.bus.Publish(log)

	return nil
}

// Currently not using WithAttrs or WithGroup. In order to satisfy the interface these are included as NO-OPs.

// WithAttrs returns a new logger with the given attributes added to the current attributes.
func (l *Logger) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Create a new Logger instance with the same options and bus
	newLogger := &Logger{
		opts: l.opts,
		bus:  l.bus,
	}

	// Copy any existing attributes from the original logger
	if len(l.attrs) > 0 {
		newLogger.attrs = make([]slog.Attr, len(l.attrs))
		copy(newLogger.attrs, l.attrs)
	}

	// Append the new attributes to the new logger's attributes
	newLogger.attrs = append(newLogger.attrs, attrs...)

	return newLogger
}

// WithGroup is not used in the current implementation and is therefore a NO-OP.
func (l *Logger) WithGroup(name string) slog.Handler { return l }

// Bus returns the underlying LogBus from the Logger.
func (l *Logger) Bus() *LogBus {
	return l.bus
}
