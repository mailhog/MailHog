package layout

import (
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/ian-kent/go-log/levels"
)

// http://logging.apache.org/log4j/1.2/apidocs/org/apache/log4j/PatternLayout.html

// DefaultTimeLayout is the default layout used by %d
var DefaultTimeLayout = "2006-01-02 15:04:05.000000000 -0700 MST"

// LegacyDefaultTimeLayout is the legacy (non-zero padded) time layout.
// Set layout.DefaultTimeLayout = layout.LegacyDefaultTimeLayout to revert behaviour.
var LegacyDefaultTimeLayout = "2006-01-02 15:04:05.999999999 -0700 MST"

type patternLayout struct {
	Layout
	Pattern string
	created int64
	re      *regexp.Regexp
}

type caller struct {
	pc       uintptr
	file     string
	line     int
	ok       bool
	pkg      string
	fullpkg  string
	filename string
}

func Pattern(pattern string) *patternLayout {
	return &patternLayout{
		Pattern: pattern,
		re:      regexp.MustCompile("%(\\w|%)(?:{([^}]+)})?"),
		created: time.Now().UnixNano(),
	}
}

func getCaller() *caller {
	pc, file, line, ok := runtime.Caller(2)

	// TODO feels nasty?
	dir, fn := filepath.Split(file)
	bits := strings.Split(dir, "/")
	pkg := bits[len(bits)-2]

	if ok {
		return &caller{pc, file, line, ok, pkg, pkg, fn}
	}
	return nil
}

func (a *patternLayout) Format(level levels.LogLevel, message string, args ...interface{}) string {

	// TODO
	// padding, e.g. %20c, %-20c, %.30c, %20.30c, %-20.30c
	// %t - thread name
	// %M - function name

	caller := getCaller()
	r := time.Now().UnixNano()

	msg := a.re.ReplaceAllStringFunc(a.Pattern, func(m string) string {
		parts := a.re.FindStringSubmatch(m)
		switch parts[1] {
		// FIXME
		// %c and %C should probably return the logger name, not the package
		// name, since that's how the logger is created in the first place!
		case "c":
			return caller.pkg
		case "C":
			return caller.pkg
		case "d":
			// FIXME specifier, e.g. %d{HH:mm:ss,SSS}
			return time.Now().Format(DefaultTimeLayout)
		case "F":
			return caller.file
		case "l":
			return fmt.Sprintf("%s/%s:%d", caller.pkg, caller.filename, caller.line)
		case "L":
			return strconv.Itoa(caller.line)
		case "m":
			return fmt.Sprintf(message, args...)
		case "n":
			// FIXME platform-specific?
			return "\n"
		case "p":
			return levels.LogLevelsToString[level]
		case "r":
			return strconv.FormatInt((r-a.created)/100000, 10)
		case "x":
			return "" // NDC
		case "X":
			return "" // MDC (must specify key)
		case "%":
			return "%"
		}
		return m
	})

	return msg
}
