package layout

/*

Layouts control the formatting of data into a printable log string.

For example, the Basic layout passes the log message and arguments
through fmt.Sprintf.

Satisfy the Layout interface to implement your own log layout.

*/

import (
	"github.com/ian-kent/go-log/levels"
)

type Layout interface {
	Format(level levels.LogLevel, message string, args ...interface{}) string
}

func Default() Layout {
	return Basic()
}
