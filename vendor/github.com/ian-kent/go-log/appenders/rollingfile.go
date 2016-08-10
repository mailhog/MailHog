package appenders

import (
	"fmt"
	"github.com/ian-kent/go-log/layout"
	"github.com/ian-kent/go-log/levels"
	"os"
	"strconv"
	"strings"
	"sync"
)

type rollingFileAppender struct {
	Appender
	layout         layout.Layout
	MaxFileSize    int64
	MaxBackupIndex int

	filename   string
	file       *os.File
	append     bool
	writeMutex sync.Mutex

	bytesWritten int64
}

func RollingFile(filename string, append bool) *rollingFileAppender {
	a := &rollingFileAppender{
		layout:         layout.Default(),
		MaxFileSize:    104857600,
		MaxBackupIndex: 1,
		append:         append,
		bytesWritten:   0,
	}
	err := a.SetFilename(filename)
	if err != nil {
		fmt.Printf("Error opening file: %s\n", err)
		return nil
	}
	return a
}

func (a *rollingFileAppender) Close() {
	if a.file != nil {
		a.file.Close()
		a.file = nil
	}
}

func (a *rollingFileAppender) Write(level levels.LogLevel, message string, args ...interface{}) {
	m := a.Layout().Format(level, message, args...)
	if !strings.HasSuffix(m, "\n") {
		m += "\n"
	}

	a.writeMutex.Lock()
	a.file.Write([]byte(m))

	a.bytesWritten += int64(len(m))
	if a.bytesWritten >= a.MaxFileSize {
		a.bytesWritten = 0
		a.rotateFile()
	}

	a.writeMutex.Unlock()
}

func (a *rollingFileAppender) Layout() layout.Layout {
	return a.layout
}

func (a *rollingFileAppender) SetLayout(layout layout.Layout) {
	a.layout = layout
}

func (a *rollingFileAppender) Filename() string {
	return a.filename
}

func (a *rollingFileAppender) SetFilename(filename string) error {
	if a.filename != filename || a.file == nil {
		a.closeFile()
		a.filename = filename
		err := a.openFile()
		return err
	}
	return nil
}

func (a *rollingFileAppender) rotateFile() {
	a.closeFile()

	lastFile := a.filename + "." + strconv.Itoa(a.MaxBackupIndex)
	if _, err := os.Stat(lastFile); err == nil {
		os.Remove(lastFile)
	}

	for n := a.MaxBackupIndex; n > 0; n-- {
		f1 := a.filename + "." + strconv.Itoa(n)
		f2 := a.filename + "." + strconv.Itoa(n+1)
		os.Rename(f1, f2)
	}

	os.Rename(a.filename, a.filename+".1")

	a.openFile()
}
func (a *rollingFileAppender) closeFile() {
	if a.file != nil {
		a.file.Close()
		a.file = nil
	}
}
func (a *rollingFileAppender) openFile() error {
	mode := os.O_WRONLY | os.O_APPEND | os.O_CREATE
	if !a.append {
		mode = os.O_WRONLY | os.O_CREATE
	}
	f, err := os.OpenFile(a.filename, mode, 0666)
	a.file = f
	return err
}
