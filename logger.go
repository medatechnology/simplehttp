package simplehttp

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal

	DEFAULT_LOG_TIME_FORMAT = "2006/01/02 15:04:05"
	DEFAULT_LOG_PREFIX      = "[MEDA] "
)

// Logger interface for all logging operations
type Logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
}

func MiddlewareLogger(log Logger) MedaMiddleware {
	return WithName("logger", SimpleLog(log))
}

// Print logs for every request (2 lines)
// [prefix] INFO [date] [time] [rid] --Started [method] [path]
// [prefix] INFO [date] [time] [rid] Completed [method] [path] [duration]
// [prefix] INFO [date] [time] [rid] Failed [method] [path] [error] [duration]
func SimpleLog(log Logger) MedaMiddlewareFunc {
	return func(next MedaHandlerFunc) MedaHandlerFunc {
		return func(c MedaContext) error {
			start := time.Now()

			// Get request ID from headers or generate new one
			requestID := c.GetHeader(HEADER_REQUEST_ID)
			if requestID == "" {
				// requestID = GenerateRequestID()
				requestID = "no-ID"
			}

			// Log request
			log.Printf("%s --Started %s %s", requestID, c.GetMethod(), c.GetPath())

			err := next(c)

			// Log response
			duration := time.Since(start)
			if err != nil {
				log.Errorf("%s Failed %s %s - %v (%s)",
					requestID, c.GetMethod(), c.GetPath(), err, duration)
			} else {
				log.Printf("%s Completed %s %s (%s)",
					requestID, c.GetMethod(), c.GetPath(), duration)
			}

			return err
		}
	}
}

// DefaultLogger holds configuration for DefaultLogger
type DefaultLogger struct {
	level  LogLevel
	logger *log.Logger
	config *DefaultLoggerConfig
}

type DefaultLoggerConfig struct {
	Level          LogLevel // this is the minimum to print out at this log
	TimeFormat     string
	Prefix         string
	Output         io.Writer
	PrintRequestID bool
}

// NewDefaultLogger creates a new DefaultLogger with optional configuration
func NewDefaultLogger(config ...*DefaultLoggerConfig) Logger {
	var cfg *DefaultLoggerConfig
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	} else {
		cfg = &DefaultLoggerConfig{
			Level:      LogLevelInfo, // this is the default, is not DEBUG
			TimeFormat: DEFAULT_LOG_TIME_FORMAT,
			Output:     os.Stdout,
			Prefix:     DEFAULT_LOG_PREFIX,
		}
	}

	return &DefaultLogger{
		logger: log.New(cfg.Output, cfg.Prefix, 0),
		level:  cfg.Level,
		config: cfg,
	}
}

func (l *DefaultLogger) formatMessage(v ...interface{}) string {
	timestamp := time.Now().Format(l.config.TimeFormat)
	// return fmt.Sprintf(" %s [%s] %s", timestamp, l.config.Prefix, fmt.Sprint(v...))
	return fmt.Sprintf(" %s %s", timestamp, fmt.Sprint(v...))
}

func (l *DefaultLogger) formatMessagef(format string, v ...interface{}) string {
	timestamp := time.Now().Format(l.config.TimeFormat)
	message := fmt.Sprintf(format, v...)
	// return fmt.Sprintf(" %s [%s] %s", timestamp, l.config.Prefix, message)
	return fmt.Sprintf(" %s %s", timestamp, message)
}

func (l *DefaultLogger) Print(v ...interface{}) {
	if l.level <= LogLevelInfo {
		l.logger.Print("INFO", l.formatMessage(v...))
	}
}

func (l *DefaultLogger) Printf(format string, v ...interface{}) {
	if l.level <= LogLevelInfo {
		l.logger.Print("INFO", l.formatMessagef(format, v...))
	}
}

func (l *DefaultLogger) Debug(v ...interface{}) {
	if l.level <= LogLevelDebug {
		l.logger.Print("DEBUG", l.formatMessage(v...))
	}
}

func (l *DefaultLogger) Debugf(format string, v ...interface{}) {
	if l.level <= LogLevelDebug {
		l.logger.Print("DEBUG", l.formatMessagef(format, v...))
	}
}

func (l *DefaultLogger) Info(v ...interface{}) {
	if l.level <= LogLevelInfo {
		l.logger.Print("INFO", l.formatMessage(v...))
	}
}

func (l *DefaultLogger) Infof(format string, v ...interface{}) {
	if l.level <= LogLevelInfo {
		l.logger.Print("INFO", l.formatMessagef(format, v...))
	}
}

func (l *DefaultLogger) Warn(v ...interface{}) {
	if l.level <= LogLevelWarn {
		l.logger.Print("WARN", l.formatMessage(v...))
	}
}

func (l *DefaultLogger) Warnf(format string, v ...interface{}) {
	if l.level <= LogLevelWarn {
		l.logger.Print("WARN", l.formatMessagef(format, v...))
	}
}

func (l *DefaultLogger) Error(v ...interface{}) {
	if l.level <= LogLevelError {
		l.logger.Print("ERROR", l.formatMessage(v...))
	}
}

func (l *DefaultLogger) Errorf(format string, v ...interface{}) {
	if l.level <= LogLevelError {
		l.logger.Print("ERROR", l.formatMessagef(format, v...))
	}
}

func (l *DefaultLogger) Fatal(v ...interface{}) {
	if l.level <= LogLevelFatal {
		l.logger.Fatal("FATAL", l.formatMessage(v...))
	}
}

func (l *DefaultLogger) Fatalf(format string, v ...interface{}) {
	if l.level <= LogLevelFatal {
		l.logger.Fatal("FATAL", l.formatMessagef(format, v...))
	}
}

// Example usage:
// logger := NewDefaultLogger(&DefaultLoggerConfig{
//     Level:      LogLevelDebug,
//     TimeFormat: "2006/01/02 15:04:05",
//     Output:     os.Stdout,
//     Prefix:     "[MEDA] ",
// })
