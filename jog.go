// Copyright 2013 Justin Wilson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package jog is used to log JSON style logging information.
// It's intended to be used with Go's log package, by calling `log.SetOutput(io.Writer)`,
// although it can be used standalone as well.
package jog

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"
)

const (
	CRITICAL = Level("critical")
	ERROR    = Level("error")
	WARNING  = Level("warning")
	INFO     = Level("info")
	DEBUG    = Level("debug")
)

// Level is the level of the data being logged
type Level string

type message struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Level   Level       `json:"level"`
	File    string      `json:"file"`
	Line    int         `json:"line"`
	Time    time.Time   `json:"time"`
}

// Logger is an interface used as the communication means for the log
type Logger interface {
	Log(p []byte) error
}

// logWRiter implements io.Writer so it can be used as log.SetOutput(logWriter)
type Jog struct {
	logger Logger
}

func (j *Jog) Log(l Level, o interface{}) error {
	_, err := j.write(newMessage(l, o))
	return err
}
func (j *Jog) Critical(o interface{}) error {
	return j.Log(CRITICAL, o)
}
func (j *Jog) Error(o interface{}) error {
	return j.Log(ERROR, o)
}
func (j *Jog) Warning(o interface{}) error {
	return j.Log(WARNING, o)
}
func (j *Jog) Info(o interface{}) error {
	return j.Log(INFO, o)
}
func (j *Jog) Debug(o interface{}) error {
	return j.Log(DEBUG, o)
}

// io.Writer
func (j *Jog) Write(p []byte) (int, error) {
	m := newMessage(INFO, nil)

	// Remove trailing "\n", added by `log.Output(int, string)`
	l := len(p) - 1
	if len(p) > 0 && p[l] == '\n' {
		p = p[0:l]
		l--
	}

	// Attempt to set JSON value of `p` and log level
	isJSONLike := l > 1 && p[0] == '{' && p[l] == '}'
	if isJSONLike && json.Unmarshal(p, &m.Data) == nil {
		m.Level = levelFrom(m.Data)
	} else {
		m.Message = string(p)
	}

	// Send to logger
	return j.write(m)
}

// Invoke the Logger with the JSON data
func (j *Jog) write(m *message) (int, error) {
	b, err := json.Marshal(m)
	if err != nil {
		m := fmt.Sprintf("[LOG FAILURE] - Failed to marshal JSON for %v, %s\n", m, err)
		os.Stderr.Write([]byte(m))
		//b = []byte(fmt.Sprintf("%#v -> %s", m, err))
	} else if err = j.logger.Log(b); err != nil {
		m := fmt.Sprintf("[LOG FAILURE] - (Logger) %s -> %s\n", err, b)
		os.Stderr.Write([]byte(m))
	}

	if err != nil {
		return 0, err
	}
	return len(b), err
}

// Pulls the `level` value from the message to be logged
func levelFrom(o interface{}) Level {
	level := INFO
	// Ensure we've got data
	m, ok := o.(map[string]interface{})
	if !ok {
		return level
	}

	// Ensure the key `level` exists
	l, ok := m["level"]
	if !ok {
		return level
	}

	// Remove `level` from data, as it exists in the message
	delete(m, "level")

	// Set the level
	switch l {
	case "critical":
		level = CRITICAL
	case "error":
		level = ERROR
	case "warning":
		level = WARNING
	case "debug":
		level = DEBUG
	}

	return level
}

func newMessage(l Level, d interface{}) *message {
	m := &message{
		Data:  d,
		Level: l,
		Time:  time.Now().UTC(),
	}
	/*m := new(message)
	m.Data = d
	m.Level = l
	m.Time = time.Now().UTC()*/

	// Set filename/line number of invoker
	ok := false
	_, m.File, m.Line, ok = runtime.Caller(4)
	if !ok {
		m.File = "???"
		m.Line = 0
	}

	return m
}

// NewWriter returns an io.Writer used to JSONify log messages
func NewWriter(l Logger) io.Writer {
	return &Jog{l}
}

// New returns a new Logger
func NewLogger(l Logger) *log.Logger {
	return log.New(&Jog{l}, "", 0)
}

func New(l Logger) *Jog {
	return &Jog{l}
}
