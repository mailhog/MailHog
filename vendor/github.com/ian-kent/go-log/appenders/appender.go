package appenders

/*

Appenders control the flow of data from a logger to an output.

For example, a Console appender outputs log data to stdout.

Satisfy the Appender interface to implement your own log appender.

*/

import (
	"github.com/ian-kent/go-log/layout"
	"github.com/ian-kent/go-log/levels"
)

type Appender interface {
	Write(level levels.LogLevel, message string, args ...interface{})
	Layout() layout.Layout
	SetLayout(layout.Layout)
}
