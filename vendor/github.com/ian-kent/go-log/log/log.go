package log

import (
	"github.com/ian-kent/go-log/levels"
	"github.com/ian-kent/go-log/logger"
	"strings"
)

var global logger.Logger

// Converts a string level (e.g. DEBUG) to a LogLevel
func Stol(level string) levels.LogLevel {
	return levels.StringToLogLevels[strings.ToUpper(level)]
}

// Returns a Logger instance
//
// If no arguments are given, the global/root logger
// instance will be returned.
//
// If at least one argument is given, the logger instance
// for that namespace will be returned.
func Logger(args ...string) logger.Logger {
	var name string
	if len(args) > 0 {
		name = args[0]
	} else {
		name = ""
	}

	if global == nil {
		global = logger.New("")
		global.SetLevel(levels.DEBUG)
	}

	l := global.GetLogger(name)

	return l
}

func Log(level levels.LogLevel, params ...interface{}) {
	Logger().Log(level, params...)
}

func Level(level levels.LogLevel)   { Logger().Level() }
func Debug(params ...interface{})   { Log(levels.DEBUG, params...) }
func Info(params ...interface{})    { Log(levels.INFO, params...) }
func Warn(params ...interface{})    { Log(levels.WARN, params...) }
func Error(params ...interface{})   { Log(levels.ERROR, params...) }
func Trace(params ...interface{})   { Log(levels.TRACE, params...) }
func Fatal(params ...interface{})   { Log(levels.FATAL, params...) }
func Printf(params ...interface{})  { Log(levels.INFO, params...) }
func Println(params ...interface{}) { Log(levels.INFO, params...) }
func Fatalf(params ...interface{})  { Log(levels.FATAL, params...) }
