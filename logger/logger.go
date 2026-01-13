// Package logger provides logging functionality for the 3x-ui panel with
// dual-backend logging (console/syslog and file) and buffered log storage for web UI.
package logger

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/mhsanaei/3x-ui/v2/config"
	"github.com/mhsanaei/3x-ui/v2/util/fastuse"
	"github.com/mhsanaei/3x-ui/v2/util/fixwrite"
	"github.com/mhsanaei/3x-ui/v2/util/ringbuffer"
	"github.com/op/go-logging"
)

const (
	logFileName     = "3xui.log"            // Log file name
	timeFormat      = "2006/01/02 15:04:05" // Log timestamp format
	logMemChunkSize = 512
	logMemLines     = 1000
)

var (
	loggingBuffer *ringbuffer.ByteRing
	logger        *logging.Logger
	logFile       *os.File
)

func init() {
	loggingBuffer = ringbuffer.NewByteRing(logMemChunkSize * logMemLines) // 500KB
}

// InitLogger initializes dual logging backends: console/syslog and file.
// Console logging uses the specified level, file logging always uses DEBUG level.
func InitLogger(level logging.Level) {
	newLogger := logging.MustGetLogger("x-ui")
	backends := make([]logging.Backend, 0, 2)

	// Console/syslog backend with configurable level
	if consoleBackend := initDefaultBackend(); consoleBackend != nil {
		leveledBackend := logging.AddModuleLevel(consoleBackend)
		leveledBackend.SetLevel(level, "x-ui")
		backends = append(backends, leveledBackend)
	}

	// File backend with DEBUG level for comprehensive logging
	if fileBackend := initFileBackend(); fileBackend != nil {
		leveledBackend := logging.AddModuleLevel(fileBackend)
		leveledBackend.SetLevel(logging.DEBUG, "x-ui")
		backends = append(backends, leveledBackend)
	}

	multiBackend := logging.MultiLogger(backends...)
	newLogger.SetBackend(multiBackend)
	logger = newLogger
}

// initDefaultBackend creates the console/syslog logging backend.
// Windows: Uses stderr directly (no syslog support)
// Unix-like: Attempts syslog, falls back to stderr
func initDefaultBackend() logging.Backend {
	var backend logging.Backend
	includeTime := false

	if runtime.GOOS == "windows" {
		// Windows: Use stderr directly (no syslog support)
		backend = logging.NewLogBackend(os.Stderr, "", 0)
		includeTime = true
	} else {
		// Unix-like: Try syslog, fallback to stderr
		if syslogBackend, err := logging.NewSyslogBackend(""); err != nil {
			fmt.Fprintf(os.Stderr, "syslog backend disabled: %v\n", err)
			backend = logging.NewLogBackend(os.Stderr, "", 0)
			includeTime = os.Getppid() > 0
		} else {
			backend = syslogBackend
		}
	}

	return logging.NewBackendFormatter(backend, newFormatter(includeTime))
}

// initFileBackend creates the file logging backend.
// Creates log directory and truncates log file on startup for fresh logs.
func initFileBackend() logging.Backend {
	logDir := config.GetLogFolder()
	if err := os.MkdirAll(logDir, 0o750); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create log folder %s: %v\n", logDir, err)
		return nil
	}

	logPath := filepath.Join(logDir, logFileName)
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o660)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open log file %s: %v\n", logPath, err)
		return nil
	}

	// Close previous log file if exists
	if logFile != nil {
		_ = logFile.Close()
	}
	logFile = file

	backend := logging.NewLogBackend(file, "", 0)
	return logging.NewBackendFormatter(backend, newFormatter(true))
}

// newFormatter creates a log formatter with optional timestamp.
func newFormatter(withTime bool) logging.Formatter {
	format := `%{level} - %{message}`
	if withTime {
		format = `%{time:` + timeFormat + `} %{level} - %{message}`
	}
	return logging.MustStringFormatter(format)
}

// CloseLogger closes the log file and cleans up resources.
// Should be called during application shutdown.
func CloseLogger() {
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
	}
}

// Debug logs a debug message and adds it to the log buffer.
func Debug(args ...any) {
	logger.Debug(args...)
	addToBuffer("DEBUG", fmt.Sprint(args...))
}

// Debugf logs a formatted debug message and adds it to the log buffer.
func Debugf(format string, args ...any) {
	logger.Debugf(format, args...)
	addToBuffer("DEBUG", fmt.Sprintf(format, args...))
}

// Info logs an info message and adds it to the log buffer.
func Info(args ...any) {
	logger.Info(args...)
	addToBuffer("INFO", fmt.Sprint(args...))
}

// Infof logs a formatted info message and adds it to the log buffer.
func Infof(format string, args ...any) {
	logger.Infof(format, args...)
	addToBuffer("INFO", fmt.Sprintf(format, args...))
}

// Notice logs a notice message and adds it to the log buffer.
func Notice(args ...any) {
	logger.Notice(args...)
	addToBuffer("NOTICE", fmt.Sprint(args...))
}

// Noticef logs a formatted notice message and adds it to the log buffer.
func Noticef(format string, args ...any) {
	logger.Noticef(format, args...)
	addToBuffer("NOTICE", fmt.Sprintf(format, args...))
}

// Warning logs a warning message and adds it to the log buffer.
func Warning(args ...any) {
	logger.Warning(args...)
	addToBuffer("WARNING", fmt.Sprint(args...))
}

// Warningf logs a formatted warning message and adds it to the log buffer.
func Warningf(format string, args ...any) {
	logger.Warningf(format, args...)
	addToBuffer("WARNING", fmt.Sprintf(format, args...))
}

// Error logs an error message and adds it to the log buffer.
func Error(args ...any) {
	logger.Error(args...)
	addToBuffer("ERROR", fmt.Sprint(args...))
}

// Errorf logs a formatted error message and adds it to the log buffer.
func Errorf(format string, args ...any) {
	logger.Errorf(format, args...)
	addToBuffer("ERROR", fmt.Sprintf(format, args...))
}

// addToBuffer adds a log entry to the in-memory ring buffer for web UI retrieval.
func addToBuffer(level string, newLog string) {
	buf := [logMemChunkSize]byte{}
	wr := fixwrite.NewFixedWriter(buf[:])

	wr.WriteString(time.Now().Format(timeFormat))
	wr.WriteString(" ")
	wr.WriteString(level)
	wr.WriteString(" - ")
	wr.WriteString(newLog)

	wr.WriteTo(loggingBuffer)
}

// GetLogs retrieves up to c log entries from the buffer that are at or below the specified level.
func GetLogs(c int, level string) []string {
	if c <= 0 {
		return nil
	}

	logs := make([]string, 0, c)
	wantLevel := string2level(level)

	b := loggingBuffer.Bytes()

	for off := 0; off+logMemChunkSize <= len(b); off += logMemChunkSize {
		chunk := fastuse.TrimZeros(b[off : off+logMemChunkSize])
		line, ok := parseLineLevel(chunk, wantLevel)
		if !ok {
			continue
		}

		if len(logs) == c {
			copy(logs, logs[1:])
			logs[c-1] = line
		} else {
			logs = append(logs, line)
		}
	}

	return logs
}

func parseLineLevel(rec []byte, maxLvl logging.Level) (line string, ok bool) {
	i := bytes.Index(rec, []byte(" - "))
	if i == -1 {
		return "", false
	}

	left := rec[:i]

	j := bytes.LastIndexByte(left, ' ')
	if j == -1 || j+1 >= len(left) {
		return "", false
	}

	recLvl, err := logging.LogLevel(fastuse.Bytes2String(left[j+1:]))
	if err != nil {
		return "", false
	}

	if levelEnabled(maxLvl, recLvl) {
		return string(rec), true
	}

	return "", false
}

func string2level(level string) logging.Level {
	switch level {
	case "debug":
		return logging.DEBUG
	case "info":
		return logging.INFO
	case "notice":
		return logging.NOTICE
	case "warning":
		return logging.WARNING
	case "err":
		return logging.ERROR
	default:
		return logging.CRITICAL
	}
}

func levelEnabled(base, level logging.Level) bool {
	return level <= base
}
