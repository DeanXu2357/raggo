package log

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

var (
	logger logr.Logger
)

func init() {
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	logger = zapr.NewLogger(zapLog)
}

// Logger returns the global logger
func Logger() logr.Logger {
	return logger
}

// SetLogger sets the global logger
func SetLogger(l logr.Logger) {
	logger = l
}

// Info logs a non-error message with the given key/value pairs as context
func Info(msg string, keysAndValues ...interface{}) {
	logger.Info(msg, keysAndValues...)
}

// Debug logs a debug message with the given key/value pairs as context
func Debug(msg string, keysAndValues ...interface{}) {
	logger.V(1).Info(msg, keysAndValues...)
}

// Error logs an error message with the given key/value pairs as context
func Error(err error, msg string, keysAndValues ...interface{}) {
	logger.Error(err, msg, keysAndValues...)
}

// V returns a logger value for a specific verbosity level
func V(level int) logr.Logger {
	return logger.V(level)
}

// WithName adds a new element to the logger's name
func WithName(name string) logr.Logger {
	return logger.WithName(name)
}

// WithValues adds some key-value pairs of context to a logger
func WithValues(keysAndValues ...interface{}) logr.Logger {
	return logger.WithValues(keysAndValues...)
}
