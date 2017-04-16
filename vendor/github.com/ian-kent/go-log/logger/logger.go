package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/ian-kent/go-log/appenders"
	"github.com/ian-kent/go-log/layout"
	"github.com/ian-kent/go-log/levels"
)

// Logger represents a logger
type Logger interface {
	Level() levels.LogLevel
	Name() string
	FullName() string
	Enabled() map[levels.LogLevel]bool
	Appender() Appender
	Children() []Logger
	Parent() Logger
	GetLogger(string) Logger
	SetLevel(levels.LogLevel)
	Log(levels.LogLevel, ...interface{})
	SetAppender(appender Appender)

	Debug(params ...interface{})
	Info(params ...interface{})
	Warn(params ...interface{})
	Error(params ...interface{})
	Trace(params ...interface{})
	Printf(params ...interface{})
	Println(params ...interface{})
	Fatal(params ...interface{})
	Fatalf(params ...interface{})
}

type logger struct {
	level       levels.LogLevel
	name        string
	enabled     map[levels.LogLevel]bool
	appender    Appender
	children    []Logger
	parent      Logger
	ExitOnFatal bool
}

// Appender represents a log appender
type Appender interface {
	Write(level levels.LogLevel, message string, args ...interface{})
	SetLayout(layout layout.Layout)
	Layout() layout.Layout
}

// New returns a new Logger
func New(name string) Logger {
	l := Logger(&logger{
		level:       levels.DEBUG,
		name:        name,
		enabled:     make(map[levels.LogLevel]bool),
		appender:    appenders.Console(),
		children:    make([]Logger, 0),
		parent:      nil,
		ExitOnFatal: true,
	})
	l.SetLevel(levels.DEBUG)
	return l
}

func unwrap(args ...interface{}) []interface{} {
	head := args[0]
	switch head.(type) {
	case func() (string, []interface{}):
		msg, args := head.(func() (string, []interface{}))()
		args = unwrap(args...)
		return append([]interface{}{msg}, args...)
	case func() []interface{}:
		args = unwrap(head.(func() []interface{})()...)
	case func(...interface{}) []interface{}:
		args = unwrap(head.(func(...interface{}) []interface{})(args[1:]...)...)
	}
	return args
}

func (l *logger) New(name string) Logger {
	lg := Logger(&logger{
		level:    levels.INHERIT,
		name:     name,
		enabled:  make(map[levels.LogLevel]bool),
		appender: nil,
		children: make([]Logger, 0),
		parent:   l,
	})
	l.children = append(l.children, lg)
	return lg
}

func (l *logger) GetLogger(name string) Logger {
	bits := strings.Split(name, ".")

	if l.name == bits[0] {
		if len(bits) == 1 {
			return l
		}

		child := bits[1]
		n := strings.Join(bits[1:], ".")
		for _, c := range l.children {
			if c.Name() == child {
				return c.GetLogger(n)
			}
		}

		lg := l.New(child)
		return lg.GetLogger(n)
	}
	lg := l.New(bits[0])
	return lg.GetLogger(name)
}

type stringer interface {
	String() string
}

func (l *logger) write(level levels.LogLevel, params ...interface{}) {
	a := l.Appender()
	if a != nil {
		if s, ok := params[0].(string); ok {
			a.Write(level, s, params[1:]...)
		} else if s, ok := params[0].(stringer); ok {
			a.Write(level, s.String(), params[1:]...)
		} else {
			a.Write(level, fmt.Sprintf("%s", params[0]), params[1:]...)
		}
	}
}

func (l *logger) Appender() Appender {
	if a := l.appender; a != nil {
		return a
	}
	if l.parent != nil {
		if a := l.parent.Appender(); a != nil {
			return a
		}
	}
	return nil
}

func (l *logger) Log(level levels.LogLevel, params ...interface{}) {
	if !l.Enabled()[level] {
		return
	}
	l.write(level, unwrap(params...)...)

	if l.ExitOnFatal && level == levels.FATAL {
		os.Exit(1)
	}
}

func (l *logger) Level() levels.LogLevel {
	if l.level == levels.INHERIT {
		return l.parent.Level()
	}
	return l.level
}

func (l *logger) Enabled() map[levels.LogLevel]bool {
	if l.level == levels.INHERIT {
		return l.parent.Enabled()
	}
	return l.enabled
}

func (l *logger) Name() string {
	return l.name
}

func (l *logger) FullName() string {
	n := l.name
	if l.parent != nil {
		p := l.parent.FullName()
		if len(p) > 0 {
			n = l.parent.FullName() + "." + n
		}
	}
	return n
}

func (l *logger) Children() []Logger {
	return l.children
}

func (l *logger) Parent() Logger {
	return l.parent
}

func (l *logger) SetLevel(level levels.LogLevel) {
	l.level = level
	for k := range levels.LogLevelsToString {
		if k <= level {
			l.enabled[k] = true
		} else {
			l.enabled[k] = false
		}
	}
}

func (l *logger) SetAppender(appender Appender) {
	l.appender = appender
}

func (l *logger) Debug(params ...interface{})   { l.Log(levels.DEBUG, params...) }
func (l *logger) Info(params ...interface{})    { l.Log(levels.INFO, params...) }
func (l *logger) Warn(params ...interface{})    { l.Log(levels.WARN, params...) }
func (l *logger) Error(params ...interface{})   { l.Log(levels.ERROR, params...) }
func (l *logger) Trace(params ...interface{})   { l.Log(levels.TRACE, params...) }
func (l *logger) Printf(params ...interface{})  { l.Log(levels.INFO, params...) }
func (l *logger) Println(params ...interface{}) { l.Log(levels.INFO, params...) }
func (l *logger) Fatal(params ...interface{})   { l.Log(levels.FATAL, params...) }
func (l *logger) Fatalf(params ...interface{})  { l.Log(levels.FATAL, params...) }
