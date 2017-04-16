package appenders

// TODO add tests

import (
	"github.com/ian-kent/go-log/layout"
	"github.com/ian-kent/go-log/levels"
	"github.com/t-k/fluent-logger-golang/fluent"
)

type fluentdAppender struct {
	Appender
	layout       layout.Layout
	fluent       *fluent.Fluent
	fluentConfig fluent.Config
}

func SafeFluentd(config fluent.Config) (*fluentdAppender, error) {
	a := &fluentdAppender{
		layout:       layout.Default(),
		fluentConfig: config,
	}
	if err := a.Open(); err != nil {
		return nil, err
	}
	return a, nil
}

func Fluentd(config fluent.Config) *fluentdAppender {
	a, _ := SafeFluentd(config)
	return a
}

func (a *fluentdAppender) Close() {
	a.fluent.Close()
	a.fluent = nil
}

func (a *fluentdAppender) Open() error {
	f, err := fluent.New(a.fluentConfig)
	if err != nil {
		return err
	}
	a.fluent = f
	return nil
}

func (a *fluentdAppender) Write(level levels.LogLevel, message string, args ...interface{}) {
	// FIXME
	// - use tag instead of "go-log"
	// - get layout to return the map
	var data = map[string]string{
		"message": a.Layout().Format(level, message, args...),
	}
	a.fluent.Post("go-log", data)
}

func (a *fluentdAppender) Layout() layout.Layout {
	return a.layout
}

func (a *fluentdAppender) SetLayout(layout layout.Layout) {
	a.layout = layout
}
