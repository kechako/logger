package logger

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

type Level int

const (
	Debug Level = iota
	Info
	Warn
	Error
	Fatal
)

const (
	tagDebug = "DEBUG: "
	tagInfo  = "INFO : "
	tagWarn  = "WARN : "
	tagError = "ERROR: "
	tagFatal = "FATAL: "
)

type Logger struct {
	debugLog *log.Logger
	infoLog  *log.Logger
	warnLog  *log.Logger
	errorLog *log.Logger
	fatalLog *log.Logger

	level Level
	depth int

	mu sync.Mutex

	closers []io.Closer
}

const defaultLogFlags = log.Ldate | log.Lmicroseconds | log.Lshortfile

var DefaultLevel = Debug

func New(opts ...Option) *Logger {
	o := options{
		level:    DefaultLevel,
		logFlags: defaultLogFlags,
	}
	for _, opt := range opts {
		opt.apply(&o)
	}

	var outputs []io.Writer

	iLogs := []io.Writer{os.Stdout}
	eLogs := []io.Writer{os.Stderr}

	if o.infoLogFile != nil {
		iLogs = append(iLogs, o.infoLogFile)
		outputs = append(outputs, o.infoLogFile)
	}
	if o.errorLogFile != nil {
		eLogs = append(eLogs, o.errorLogFile)
		outputs = append(outputs, o.errorLogFile)
	}

	l := &Logger{
		debugLog: log.New(io.MultiWriter(iLogs...), tagDebug, o.logFlags),
		infoLog:  log.New(io.MultiWriter(iLogs...), tagInfo, o.logFlags),
		warnLog:  log.New(io.MultiWriter(eLogs...), tagWarn, o.logFlags),
		errorLog: log.New(io.MultiWriter(eLogs...), tagError, o.logFlags),
		fatalLog: log.New(io.MultiWriter(eLogs...), tagFatal, o.logFlags),
		level:    o.level,
	}

	for _, output := range outputs {
		if c, ok := output.(io.Closer); ok {
			l.closers = append(l.closers, c)
		}
	}

	return l
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var hasErr bool
	for _, c := range l.closers {
		if err := c.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close log %v: %v\n", c, err)
			hasErr = true
		}
	}

	if hasErr {
		return errors.New("failed to close some logs")
	}

	return nil
}

func (l *Logger) SetDepth(depth int) {
	if depth < 0 {
		panic("depth must be more than or equal to 0")
	}

	l.depth = depth
}

func (l *Logger) Depth() int {
	return l.depth
}

func (l *Logger) log(level Level, text string) {
	if level < l.level {
		return
	}

	l.mu.Lock()

	switch level {
	case Debug:
		l.debugLog.Output(3+l.depth, text)
	case Info:
		l.infoLog.Output(3+l.depth, text)
	case Warn:
		l.warnLog.Output(3+l.depth, text)
	case Error:
		l.errorLog.Output(3+l.depth, text)
	case Fatal:
		l.fatalLog.Output(3+l.depth, text)
	}

	l.mu.Unlock()
}

func (l *Logger) Debug(v ...interface{}) {
	l.log(Debug, fmt.Sprint(v...))
}

func (l *Logger) Debugln(v ...interface{}) {
	l.log(Debug, fmt.Sprintln(v...))
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.log(Debug, fmt.Sprintf(format, v...))
}

func (l *Logger) Info(v ...interface{}) {
	l.log(Info, fmt.Sprint(v...))
}

func (l *Logger) Infoln(v ...interface{}) {
	l.log(Info, fmt.Sprintln(v...))
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.log(Info, fmt.Sprintf(format, v...))
}

func (l *Logger) Warn(v ...interface{}) {
	l.log(Warn, fmt.Sprint(v...))
}

func (l *Logger) Warnln(v ...interface{}) {
	l.log(Warn, fmt.Sprintln(v...))
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.log(Warn, fmt.Sprintf(format, v...))
}

func (l *Logger) Error(v ...interface{}) {
	l.log(Error, fmt.Sprint(v...))
}

func (l *Logger) Errorln(v ...interface{}) {
	l.log(Error, fmt.Sprintln(v...))
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.log(Error, fmt.Sprintf(format, v...))
}

func (l *Logger) Fatal(v ...interface{}) {
	l.log(Fatal, fmt.Sprint(v...))
	l.Close()
	os.Exit(1)
}

func (l *Logger) Fatalln(v ...interface{}) {
	l.log(Fatal, fmt.Sprintln(v...))
	l.Close()
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.log(Fatal, fmt.Sprintf(format, v...))
	l.Close()
	os.Exit(1)
}

type options struct {
	level        Level
	infoLogFile  io.Writer
	errorLogFile io.Writer
	logFlags     int
}

type Option interface {
	apply(o *options)
}

type OptionFunc func(o *options)

func (f OptionFunc) apply(o *options) {
	f(o)
}

func WithLevel(level Level) Option {
	return OptionFunc(func(o *options) {
		o.level = level
	})
}

func WithInfoLogFile(logFile io.Writer) Option {
	return OptionFunc(func(o *options) {
		o.infoLogFile = logFile
	})
}

func WithErrorLogFile(logFile io.Writer) Option {
	return OptionFunc(func(o *options) {
		o.errorLogFile = logFile
	})
}

func WithLogFlags(flags int) Option {
	return OptionFunc(func(o *options) {
		o.logFlags = flags
	})
}
