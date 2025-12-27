package logger

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const lastNSegments = 2

// New creates and returns a new logger with the specified log level and format (json or text).
// Panics if an invalid format is provided.
func New(lvl, format string) *slog.Logger {
	opts := &slog.HandlerOptions{AddSource: true, Level: logLevel(lvl), ReplaceAttr: replaceAttr}
	var handler slog.Handler
	switch format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		panic("invalid logging format, use json or text")
	}

	return slog.New(&ContextHandler{Handler: handler})
}

// logLevel maps a string log level to a corresponding slog.Level. Defaults to slog.LevelInfo for unrecognized levels.
func logLevel(lvl string) slog.Level {
	lvl = strings.ToUpper(lvl)
	switch lvl {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// replaceAttr modifies slog.Attr instances, altering the `Value` of the `SourceKey` to show only the last segments of the path.
func replaceAttr(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.SourceKey {
		tmp := strings.Split(a.Value.String(), string(filepath.Separator))
		n := lastNSegments
		if len(tmp) < n {
			n = len(tmp)
		}
		a.Value = slog.StringValue(strings.Join(tmp[len(tmp)-n:], "/"))
	}
	return a
}
