package logger

import (
	"fmt"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/configuration"
	"io"
	"log"
	"os"
	"path/filepath"
)

type Logger struct {
	level LogLevel
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
	fatal *log.Logger
}

func New(configs configuration.LoggerConf) *Logger {
	flags := 0

	var err error
	if err = os.MkdirAll(filepath.Dir(configs.File), 0755); err != nil {
		log.Fatalln(fmt.Errorf("failed to create log directory: %w", err))
	}

	var logFile *os.File
	if logFile, err = os.OpenFile(configs.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		log.Fatalln(fmt.Errorf("failed to open log file: %w", err))
	}

	defaultWriter := io.MultiWriter(os.Stderr, logFile)
	errWriter := io.MultiWriter(os.Stderr, logFile)

	return &Logger{
		level: ParseStatus(configs.Level),
		debug: log.New(defaultWriter, "DEBUG: ", flags),
		info:  log.New(defaultWriter, "INFO: ", flags),
		warn:  log.New(defaultWriter, "WARN: ", flags),
		error: log.New(errWriter, "ERROR: ", flags),
		fatal: log.New(errWriter, "FATAL: ", flags),
	}
}

func (l *Logger) Debug(msg string) {
	if l.level <= DEBUG {
		l.debug.Println(msg)
	}
}

func (l *Logger) Info(msg string) {
	if l.level <= INFO {
		l.info.Println(msg)
	}
}

func (l *Logger) Warn(msg string) {
	if l.level <= WARN {
		l.warn.Println(msg)
	}
}

func (l *Logger) Error(msg string) {
	if l.level <= ERROR {
		l.error.Println(msg)
	}
}

func (l *Logger) Fatal(msg string) {
	if l.level <= FATAL {
		l.fatal.Fatalln(msg)
	}
}
