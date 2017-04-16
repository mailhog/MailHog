package levels

type LogLevel int

const (
	FATAL LogLevel = iota
	ERROR
	INFO
	WARN
	DEBUG
	TRACE
	INHERIT
)

var StringToLogLevels = map[string]LogLevel{
	"TRACE":   TRACE,
	"DEBUG":   DEBUG,
	"WARN":    WARN,
	"INFO":    INFO,
	"ERROR":   ERROR,
	"FATAL":   FATAL,
	"INHERIT": INHERIT,
}

var LogLevelsToString = map[LogLevel]string{
	TRACE:   "TRACE",
	DEBUG:   "DEBUG",
	WARN:    "WARN",
	INFO:    "INFO",
	ERROR:   "ERROR",
	FATAL:   "FATAL",
	INHERIT: "INHERIT",
}
