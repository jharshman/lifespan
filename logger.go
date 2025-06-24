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
	opts Options
	bus  *LogBus
}

// Options encapsulates options for the logger.
type Options struct {
	Level slog.Leveler
}

// New returns a pointer to a Logger. If the provided bsize is less than the defaultBufferSize, it will default to the
// defaultBufferSize.
func New(bsize int64, opts *Options) *Logger {
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
func (l *Logger) Enabled(_ context.Context, level slog.Level) bool {
	return level >= l.opts.Level.Level()
}

// Handle takes log records and sends them to the underlying LogBus.
func (l *Logger) Handle(_ context.Context, r slog.Record) error {

	log := Log{
		Msg:      r.Message,
		Level:    r.Level.String(),
		Metadata: make(map[string]any, r.NumAttrs()),
	}

	// set timestamp in UTC
	if !r.Time.IsZero() {
		log.Timestamp = r.Time.UTC()
	}

	// extract attributes from slog.Record to finish created Log.
	// Attrs will loop over each attribute in the slog.Record unless false is returned.
	// Here we will extract job_id and group_id and then set any additional metadata into the Log.Metadata field.
	r.Attrs(func(attr slog.Attr) bool {
		v := attr.Value.Resolve()
		if attr.Equal(slog.Attr{}) {
			return true
		}
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

func (l *Logger) WithAttrs(attrs []slog.Attr) slog.Handler { return l }
func (l *Logger) WithGroup(name string) slog.Handler       { return l }
